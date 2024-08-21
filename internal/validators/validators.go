package validators

import (
	"net/http"
	"slices"
	"strconv"
)

func allowedMetricTypes() []string {
	return []string{"gauge", "counter"}
}

func TypesMetrics() map[string][]string {
	return map[string][]string{
		"gauge": []string{
			"Alloc", "BuckHashSys", "Frees", "GCCPUFraction",
			"GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects",
			"HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse",
			"MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC",
			"NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs",
			"StackInuse", "StackSys", "Sys", "TotalAlloc",
		},
		"counter": []string{"PollCount"},
	}
}

func ValidateMetricType(metricType string, res http.ResponseWriter) {
	if !slices.Contains(allowedMetricTypes(), metricType) {
		http.Error(res, "Metric type is not supported!", http.StatusBadRequest)
		return
	}
}

func ValidateMetricName(metricType string, metricName string, res http.ResponseWriter) {
	if !slices.Contains(TypesMetrics()[metricType], metricName) {
		http.Error(res, "Metric not exists!", http.StatusNotFound)
		return
	}
}

func ValidateMetricValue(metricType string, metricValue string, res http.ResponseWriter) {
	if metricType == "gauge" {
		_, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(res, "Metric value not supported!", http.StatusBadRequest)
			return
		}
	}

	if metricType == "counter" {
		_, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(res, "Metric value not supported!", http.StatusBadRequest)
			return
		}
	}
}
