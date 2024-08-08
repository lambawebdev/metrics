package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"

	"github.com/lambawebdev/metrics/internal/storage"
)

func allowedMetricTypes() []string {
	return []string{"gauge", "counter"}
}

func UpdateMetric(res http.ResponseWriter, req *http.Request) {
	storage := new(storage.MemStorage)
	storage.GaugeMetric = make(map[string]float64)
	storage.CounterMetric = make(map[string]int64)

	if !slices.Contains(allowedMetricTypes(), req.PathValue("type")) {
		http.Error(res, "Metric type is not supported!", http.StatusBadRequest)
		return
	}

	if req.PathValue("type") == "gauge" {
		value, err := strconv.ParseFloat(req.PathValue("value"), 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
		}

		storage.AddGauge(req.PathValue("name"), value)
	}

	if req.PathValue("type") == "counter" {
		value, err := strconv.ParseInt(req.PathValue("value"), 10, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
		}

		storage.AddCounter(req.PathValue("name"), value)
	}

	res.Header().Set("content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
}

func SendMetric(metricType string, metricName string, metricValue interface{}) {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/update/%s/%s/%v", metricType, metricName, metricValue), nil)
	if err != nil {
		log.Fatalln(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		os.Exit(1)
	}

	defer res.Body.Close()
}
