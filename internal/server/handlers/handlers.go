package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/lambawebdev/metrics/internal/models"
	"github.com/lambawebdev/metrics/internal/server/storage"
	"github.com/lambawebdev/metrics/internal/validators"
)

type MetricHandler struct {
	storage storage.MetricStorage
}

type MetricHandlerInterface interface {
	GetMetric(res http.ResponseWriter, req *http.Request)
	GetMetricV2(res http.ResponseWriter, req *http.Request)
	GetMetrics(res http.ResponseWriter)
	UpdateMetric(res http.ResponseWriter, req *http.Request)
	UpdateMetricV2(res http.ResponseWriter, req *http.Request)
	Ping(res http.ResponseWriter, db *sql.DB)
}

func NewMetricHandler(storage storage.MetricStorage) *MetricHandler {
	return &MetricHandler{
		storage: storage,
	}
}

func (mh *MetricHandler) GetMetric(res http.ResponseWriter, req *http.Request) {
	metricType := req.PathValue("type")
	metricName := req.PathValue("name")

	validators.ValidateMetricType(metricType, res)
	metric, found := mh.storage.GetMetric(metricName, metricType)

	if !found {
		http.Error(res, "Metric not exists!", http.StatusNotFound)
		return
	}

	var value interface{}

	if metricType == "gauge" {
		value = metric.Value
	}

	if metricType == "counter" {
		value = metric.Delta
	}

	res.Header().Set("content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(value)
}

func (mh *MetricHandler) GetMetrics(res http.ResponseWriter) {
	metricsValues := mh.storage.GetAll()

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

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &m); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	m, _ = mh.storage.GetMetric(m.ID, m.MType)

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
		mh.storage.AddGauge(metricName, value)
	}

	if metricType == "counter" {
		value, _ := strconv.ParseInt(metricValue, 10, 64)
		mh.storage.AddCounter(metricName, value)
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

		mh.storage.AddGauge(m.ID, *m.Value)
	}

	if m.MType == "counter" {
		if m.Delta == nil {
			http.Error(res, "delta have to be present", http.StatusBadRequest)
			return
		}
		mh.storage.AddCounter(m.ID, *m.Delta)
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

func (mh *MetricHandler) Ping(res http.ResponseWriter, db *sql.DB) {
	if err := db.Ping(); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	res.WriteHeader(http.StatusOK)
}

func (mh *MetricHandler) UpdateMetricBatch(res http.ResponseWriter, req *http.Request) {
	var buf bytes.Buffer

	var metrics []models.Metrics

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &metrics); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	mh.storage.AddBatch(metrics)

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
}
