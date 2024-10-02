package report

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/lambawebdev/metrics/internal/agent/config"
	"github.com/lambawebdev/metrics/internal/models"
	"github.com/lambawebdev/metrics/internal/validators"
	"github.com/shirou/gopsutil/v4/mem"
)

type Monitor struct {
	Alloc           uint64
	BuckHashSys     uint64
	Frees           uint64
	GCCPUFraction   float64
	GCSys           uint64
	HeapAlloc       uint64
	HeapIdle        uint64
	HeapInuse       uint64
	HeapObjects     uint64
	HeapReleased    uint64
	HeapSys         uint64
	LastGC          uint64
	Lookups         uint64
	MCacheInuse     uint64
	MCacheSys       uint64
	MSpanInuse      uint64
	MSpanSys        uint64
	Mallocs         uint64
	NextGC          uint64
	NumForcedGC     uint32
	NumGC           uint32
	OtherSys        uint64
	PauseTotalNs    uint64
	StackInuse      uint64
	StackSys        uint64
	Sys             uint64
	TotalAlloc      uint64
	PollCount       uint64
	RandomValue     uint64
	TotalMemory     float64
	FreeMemory      float64
	CPUutilization1 float64
}

func Start() {
	var m Monitor

	//32 метрики всего
	ch := make(chan models.Metrics, 32)

	pollTicker := time.NewTicker(time.Duration(config.GetFlagPollIntervalSeconds()) * time.Second)
	defer pollTicker.Stop()

	reportTicker := time.NewTicker(time.Duration(config.GetFlagReportIntervalSeconds()) * time.Second)
	defer reportTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			m = GetRuntimeMetrics(m)
			m = GetAdditionalMetrics(m)
		case <-reportTicker.C:
			SendMetrics(m, ch)
		}
	}
}

func GetRuntimeMetrics(m Monitor) Monitor {
	var rtm runtime.MemStats

	runtime.ReadMemStats(&rtm)
	m.Alloc = rtm.Alloc
	m.BuckHashSys = rtm.BuckHashSys
	m.Frees = rtm.Frees
	m.GCCPUFraction = rtm.GCCPUFraction
	m.GCSys = rtm.GCSys
	m.HeapAlloc = rtm.HeapAlloc
	m.HeapIdle = rtm.HeapIdle
	m.HeapInuse = rtm.HeapInuse
	m.HeapObjects = rtm.HeapObjects
	m.HeapReleased = rtm.HeapReleased
	m.HeapSys = rtm.HeapSys
	m.LastGC = rtm.LastGC
	m.Lookups = rtm.Lookups
	m.MCacheInuse = rtm.MCacheInuse
	m.MCacheSys = rtm.MCacheSys
	m.MSpanInuse = rtm.MSpanInuse
	m.MSpanSys = rtm.MSpanSys
	m.Mallocs = rtm.Mallocs
	m.NextGC = rtm.NextGC
	m.NumForcedGC = rtm.NumForcedGC
	m.NumGC = rtm.NumGC
	m.OtherSys = rtm.OtherSys
	m.PauseTotalNs = rtm.PauseTotalNs
	m.StackInuse = rtm.StackInuse
	m.StackSys = rtm.StackSys
	m.Sys = rtm.Sys
	m.TotalAlloc = rtm.TotalAlloc
	m.RandomValue = rand.Uint64()

	m.PollCount++

	return m
}

func GetAdditionalMetrics(m Monitor) Monitor {
	v, _ := mem.VirtualMemory()

	m.TotalMemory = float64(v.Total)
	m.FreeMemory = float64(v.Free)
	m.CPUutilization1 = float64(v.HugePagesFree)

	return m
}

func SendMetrics(m Monitor, ch chan models.Metrics) {
	//итератор по m структуре и записываем в канал
	writeMetricToChannel(m, ch)

	//worker pool 2 go routine и в каждую go рутину передаем канал
	for w := uint64(1); w <= config.GetWorkerPoolsLimit(); w++ {
		go worker(w, ch)
	}
}

func SendMetricsBatch(m Monitor) {
	metrics := prepareMetrics(m)
	sendMetricsBatchReq(metrics)
}

func worker(id uint64, metrics <-chan models.Metrics) {
	fmt.Println("woker", id)
	for metric := range metrics {
		for _, backoff := range backoffSchedule {
			fmt.Println(metric.Delta, metric.Value)
			err := sendMetricReq(metric)

			if err == nil {
				break
			}

			fmt.Fprintf(os.Stderr, "Request error: %+v\n", err)
			fmt.Fprintf(os.Stderr, "Retrying in %v\n", backoff)
			time.Sleep(backoff)
		}
	}
}

func writeMetricToChannel(m Monitor, ch chan<- models.Metrics) {
	var monitor = reflect.ValueOf(m)

	for metricType, metrics := range validators.TypesMetrics() {
		for _, metricName := range metrics {
			metricValue := reflect.Indirect(monitor).FieldByName(metricName)

			metrics := models.Metrics{}
			metrics.ID = metricName
			metrics.MType = metricType

			k := metricValue.Kind()

			if metricType == "gauge" {
				if k == reflect.Uint32 {
					value := float64(metricValue.Interface().(uint32))
					metrics.Value = &value
				}

				if k == reflect.Uint64 {
					value := float64(metricValue.Interface().(uint64))
					metrics.Value = &value
				}

				if k == reflect.Float64 {
					value := float64(metricValue.Interface().(float64))
					metrics.Value = &value
				}
			}

			if metricType == "counter" {
				value := int64(metricValue.Interface().(uint64))
				metrics.Delta = &value
			}

			ch <- metrics
		}
	}
}

func prepareMetrics(m Monitor) []models.Metrics {
	var monitor = reflect.ValueOf(m)
	var batch []models.Metrics

	for metricType, metrics := range validators.TypesMetrics() {
		for _, metricName := range metrics {
			metricValue := reflect.Indirect(monitor).FieldByName(metricName)

			metrics := models.Metrics{}
			metrics.ID = metricName
			metrics.MType = metricType

			k := metricValue.Kind()

			if metricType == "gauge" {
				if k == reflect.Uint32 {
					value := float64(metricValue.Interface().(uint32))
					metrics.Value = &value
				}

				if k == reflect.Uint64 {
					value := float64(metricValue.Interface().(uint64))
					metrics.Value = &value
				}

				if k == reflect.Float64 {
					value := float64(metricValue.Interface().(float64))
					metrics.Value = &value
				}
			}

			if metricType == "counter" {
				value := int64(metricValue.Interface().(uint64))
				metrics.Delta = &value
			}

			batch = append(batch, metrics)
		}
	}

	return batch
}

var client = resty.New()

func sendMetricReq(metrics models.Metrics) error {
	body, err := json.Marshal(metrics)

	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s/update/", config.GetFlagRunAddr())

	request := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(body)

	if config.GetSecretKey() != "" {
		secretKey := []byte(config.GetSecretKey())
		hmac, err := getHmacBody(body, secretKey)

		if err != nil {
			return err
		}

		request.SetHeader("HashSHA256", hmac)
	}

	_, err = request.Post(url)
	if err != nil {
		return err
	}

	return nil
}

func sendMetricsBatchReq(metrics []models.Metrics) error {
	body, err := json.Marshal(metrics)

	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s/updates/", config.GetFlagRunAddr())

	request := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(body)

	if config.GetSecretKey() != "" {
		secretKey := []byte(config.GetSecretKey())
		hmac, err := getHmacBody(body, secretKey)

		if err != nil {
			return err
		}

		request.SetHeader("HashSHA256", hmac)
	}

	_, err = request.Post(url)
	if err != nil {
		return err
	}

	return nil
}

var backoffSchedule = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

func getHmacBody(msg []byte, key []byte) (string, error) {
	hmac := hmac.New(sha256.New, key)
	_, err := hmac.Write(msg)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hmac.Sum(nil)), nil
}
