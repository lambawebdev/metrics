package main

import (
	"reflect"
	"runtime"
	"time"

	"github.com/lambawebdev/metrics/internal/config"
	"github.com/lambawebdev/metrics/internal/handlers"
	"github.com/lambawebdev/metrics/internal/validators"
)

type Monitor struct {
	Alloc         uint64
	BuckHashSys   uint64
	Frees         uint64
	GCCPUFraction float64
	GCSys         uint64
	HeapAlloc     uint64
	HeapIdle      uint64
	HeapInuse     uint64
	HeapObjects   uint64
	HeapReleased  uint64
	HeapSys       uint64
	LastGC        uint64
	Lookups       uint64
	MCacheInuse   uint64
	MCacheSys     uint64
	MSpanInuse    uint64
	MSpanSys      uint64
	Mallocs       uint64
	NextGC        uint64
	NumForcedGC   uint32
	NumGC         uint32
	OtherSys      uint64
	PauseTotalNs  uint64
	StackInuse    uint64
	StackSys      uint64
	Sys           uint64
	TotalAlloc    uint64
	PollCount     uint64
}

func NewMonitor(m *Monitor) {
	m.PollCount = 0
	var rtm runtime.MemStats
	var interval = time.Duration(config.GetFlagPollIntervalSeconds()) * time.Second
	for {
		<-time.After(interval)

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

		m.PollCount += 1
	}
}

func Report(m *Monitor) {
	var interval = time.Duration(config.GetFlagReportIntervalSeconds()) * time.Second
	var monitor = reflect.ValueOf(m)

	for {
		<-time.After(interval)

		for metricType, metrics := range validators.TypesMetrics() {
			for _, metricName := range metrics {
				metricValue := reflect.Indirect(monitor).FieldByName(metricName)
				handlers.SendMetric(metricType, metricName, metricValue)
			}
		}
	}
}

func main() {
	config.ParseFlags()

	var m Monitor

	go NewMonitor(&m)
	Report(&m)
}
