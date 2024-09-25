package storage

import "github.com/lambawebdev/metrics/internal/models"

type MetricStorage interface {
	AddGauge(metricName string, metricValue float64)
	AddCounter(metricName string, metricValue int64)
	GetMetric(metricName string, metricType string) (models.Metrics, bool)
	GetAll() []models.Metrics
	AddBatch(metrics []models.Metrics)
}
