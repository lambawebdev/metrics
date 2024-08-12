package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/lambawebdev/metrics/internal/storage"
	"github.com/lambawebdev/metrics/internal/validators"
)

func GetMetrics(res http.ResponseWriter, req *http.Request, storage *storage.MemStorage) {
	metricsValues := storage.GetAll()

	res.Header().Set("content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(metricsValues)
}

func GetMetric(res http.ResponseWriter, req *http.Request, storage *storage.MemStorage) {
	metricType := req.PathValue("type")
	metricName := req.PathValue("name")

	validators.ValidateMetricType(metricType, res)
	metricValue := storage.GetMetricValue(metricName)

	log.Print(metricValue)

	if metricValue == nil {
		http.Error(res, "Metric not exists!", http.StatusNotFound)
		return
	}

	res.Header().Set("content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(metricValue)
}

func UpdateMetric(res http.ResponseWriter, req *http.Request, storage *storage.MemStorage) {
	metricType := req.PathValue("type")
	metricName := req.PathValue("name")
	metricValue := req.PathValue("value")

	validators.ValidateMetricType(metricType, res)
	validators.ValidateMetricValue(metricType, metricValue, res)

	if metricType == "gauge" {
		value, _ := strconv.ParseFloat(metricValue, 64)
		storage.AddGauge(metricName, value)
	}

	if metricType == "counter" {
		value, _ := strconv.ParseInt(metricValue, 10, 64)
		storage.AddCounter(metricName, value)
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
