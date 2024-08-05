package main

import (
	"net/http"
	"strconv"
	"slices"
)

func allowedMetricTypes() []string {
	return []string{"gauge", "counter"}
}

func allowedMetricNames() []string {
	return []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction",
		"GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects",
		"HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse",
		"MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC",
		"NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse",
		"StackInuse", "StackSys", "Sys", "TotalAlloc",
	}
}

type Metric interface {
	AddGauge(metricType string, metricValue float64)
	AddCounter(metricType string, metricValue int64)
}

type MemStorage struct {
	GaugeMetric   map[string]float64
	CounterMetric map[string]int64
}

func (u *MemStorage) AddGauge(metricType string, metricValue float64) {
	u.GaugeMetric[metricType] = metricValue
}

func (u *MemStorage) AddCounter(metricType string, metricValue int64) {
	u.CounterMetric[metricType] += metricValue
}

func updateMetric(res http.ResponseWriter, req *http.Request) {
	if !slices.Contains(allowedMetricTypes(), req.PathValue("type")) {
		http.Error(res, "Metric type is not supported!", http.StatusBadRequest)
		return
	}

	if req.PathValue("type") == "gauge" {
		_, err := strconv.ParseInt(req.PathValue("value"), 10, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
		}
	}

	if req.PathValue("type") == "counter" {
		_, err := strconv.ParseFloat(req.PathValue("value"), 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
		}
	}

	res.WriteHeader(http.StatusOK)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`POST /update/{type}/{name}/{value}`, updateMetric)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
