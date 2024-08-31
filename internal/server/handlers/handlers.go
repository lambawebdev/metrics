package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/lambawebdev/metrics/internal/models"
	"github.com/lambawebdev/metrics/internal/server/storage"
	"github.com/lambawebdev/metrics/internal/validators"
)

func GetMetrics(res http.ResponseWriter, storage *storage.MemStorage) {
	metricsValues := storage.GetAll()

	res.Header().Set("Content-Type", "text/html")
	res.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(res).Encode(metricsValues); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
}

func GetMetricV2(res http.ResponseWriter, req *http.Request, storage *storage.MemStorage) {
	var m models.Metrics
	var buf bytes.Buffer

	tempValue := float64(0)
	tempDelta := int64(0)

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &m); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	metricValue := storage.GetMetricValue(m.ID)

	if metricValue == nil && m.MType == "gauge" {
		m.Value = &tempValue
	}

	if metricValue == nil && m.MType == "counter" {
		m.Delta = &tempDelta
	}

	if metricValue != nil {
		if m.MType == "gauge" {
			value := metricValue.(float64)
			m.Value = &value
		}

		if m.MType == "counter" {
			value := metricValue.(int64)
			m.Delta = &value
		}
	}

	resp, err := json.Marshal(m)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}

func GetMetric(res http.ResponseWriter, req *http.Request, storage *storage.MemStorage) {
	metricType := req.PathValue("type")
	metricName := req.PathValue("name")

	validators.ValidateMetricType(metricType, res)
	metricValue := storage.GetMetricValue(metricName)

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

func UpdateMetricV2(res http.ResponseWriter, req *http.Request, storage *storage.MemStorage) {
	var m models.Metrics
	var buf bytes.Buffer

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &m); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if !validators.ValidateMetricType(m.MType, res) {
		return
	}

	if m.MType == "gauge" {
		if m.Value == nil {
			http.Error(res, "value have to be present", http.StatusBadRequest)
			return
		}

		storage.AddGauge(m.ID, *m.Value)
	}

	if m.MType == "counter" {
		if m.Delta == nil {
			http.Error(res, "delta have to be present", http.StatusBadRequest)
			return
		}
		storage.AddCounter(m.ID, *m.Delta)
	}

	resp, err := json.Marshal(m)

	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}
