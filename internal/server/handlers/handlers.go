package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"

	"github.com/lambawebdev/metrics/internal/models"
	"github.com/lambawebdev/metrics/internal/server/storage"
	"github.com/lambawebdev/metrics/internal/validators"
)

type MetricHandler struct {
	memStorage *storage.MemStorage
}

type MetricHandlerInterface interface {
	GetMetric(res http.ResponseWriter, req *http.Request)
	GetMetricV2(res http.ResponseWriter, req *http.Request)
	GetMetrics(res http.ResponseWriter)
	UpdateMetric(res http.ResponseWriter, req *http.Request)
	UpdateMetricV2(res http.ResponseWriter, req *http.Request)
}

func NewMetricHandler(memStorage *storage.MemStorage) *MetricHandler {
	return &MetricHandler{
		memStorage: memStorage,
	}
}

func (mh *MetricHandler) GetMetric(res http.ResponseWriter, req *http.Request) {
	metricType := req.PathValue("type")
	metricName := req.PathValue("name")

	validators.ValidateMetricType(metricType, res)
	metricValue := mh.memStorage.GetMetricValue(metricName)

	if metricValue == nil {
		http.Error(res, "Metric not exists!", http.StatusNotFound)
		return
	}

	res.Header().Set("content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(metricValue)
}

func (mh *MetricHandler) GetMetrics(res http.ResponseWriter) {
	metricsValues := mh.memStorage.GetAll()

	res.Header().Set("Content-Type", "text/html")
	res.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(res).Encode(metricsValues); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
}

func (mh *MetricHandler) GetMetricV2(res http.ResponseWriter, req *http.Request) {
	var m models.Metrics
	var buf bytes.Buffer

	defValue := float64(0)
	defDelta := int64(0)

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &m); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	metricValue := mh.memStorage.GetMetricValue(m.ID)

	if metricValue == nil && m.MType == "gauge" {
		m.Value = &defValue
	}

	if metricValue == nil && m.MType == "counter" {
		m.Delta = &defDelta
	}

	if metricValue != nil {
		if m.MType == "gauge" {
			if reflect.TypeOf(metricValue).Kind() == reflect.Float64 {
				defValue = float64(metricValue.(float64))
			}

			if reflect.TypeOf(metricValue).Kind() == reflect.Int64 {
				defValue = float64(metricValue.(int64))
			}

			m.Value = &defValue
		}

		if m.MType == "counter" {
			if reflect.TypeOf(metricValue).Kind() == reflect.Float64 {
				defDelta = int64(metricValue.(float64))
			}

			if reflect.TypeOf(metricValue).Kind() == reflect.Int64 {
				defDelta = int64(metricValue.(int64))
			}

			m.Delta = &defDelta
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

func (mh *MetricHandler) UpdateMetric(res http.ResponseWriter, req *http.Request) {
	metricType := req.PathValue("type")
	metricName := req.PathValue("name")
	metricValue := req.PathValue("value")

	validators.ValidateMetricType(metricType, res)
	validators.ValidateMetricValue(metricType, metricValue, res)

	if metricType == "gauge" {
		value, _ := strconv.ParseFloat(metricValue, 64)
		mh.memStorage.AddGauge(metricName, value)
	}

	if metricType == "counter" {
		value, _ := strconv.ParseInt(metricValue, 10, 64)
		mh.memStorage.AddCounter(metricName, value)
	}

	res.Header().Set("content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
}

func (mh *MetricHandler) UpdateMetricV2(res http.ResponseWriter, req *http.Request) {
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

		mh.memStorage.AddGauge(m.ID, *m.Value)
	}

	if m.MType == "counter" {
		if m.Delta == nil {
			http.Error(res, "delta have to be present", http.StatusBadRequest)
			return
		}
		mh.memStorage.AddCounter(m.ID, *m.Delta)
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
