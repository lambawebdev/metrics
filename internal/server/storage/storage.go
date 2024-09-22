package storage

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/lambawebdev/metrics/internal/models"
	"github.com/lambawebdev/metrics/internal/server/config"
)

type MetricStorage interface {
	AddGauge(metricName string, metricValue float64)
	AddCounter(metricName string, metricValue int64)
	GetMetric(metricName string, metricType string) (models.Metrics, bool)
	GetAll() []models.Metrics
	AddBatch(metrics []models.Metrics)
}

type MemStorage struct {
	Metrics []models.Metrics
}

func GetStorageFactory(db *sql.DB) (MetricStorage, error) {
	if databaseDsn := os.Getenv("DATABASE_DSN"); databaseDsn != "" {
		return NewPGSQLMetricRepository(db), nil
	}

	return InitMemStorage(), nil
}

func (u *MemStorage) AddGauge(metricName string, metricValue float64) {
	var metric models.Metrics
	metric.MType = "gauge"
	metric.ID = metricName
	metric.Value = &metricValue

	for index, m := range u.Metrics {
		if m.ID == metricName && m.MType == "gauge" {
			u.Metrics = append(u.Metrics[:index], u.Metrics[index+1:]...)
		}
	}

	u.Metrics = append(u.Metrics, metric)
}

func (u *MemStorage) AddCounter(metricName string, metricValue int64) {
	var metric models.Metrics
	metric.MType = "counter"
	metric.ID = metricName
	metric.Delta = &metricValue

	for index, m := range u.Metrics {
		if m.ID == metricName && m.MType == "counter" {
			v := metricValue + *m.Delta
			metric.Delta = &v

			u.Metrics = append(u.Metrics[:index], u.Metrics[index+1:]...)
		}
	}

	u.Metrics = append(u.Metrics, metric)
}

func (u *MemStorage) GetMetric(metricName string, metricType string) (models.Metrics, bool) {
	for _, metric := range u.Metrics {
		if metric.ID == metricName {
			return metric, true
		}
	}

	var m models.Metrics

	defValue := float64(0)
	defDelta := int64(0)

	m.ID = metricName
	m.MType = metricType

	if metricType == "gauge" {
		m.Value = &defValue
	}

	if metricType == "counter" {
		m.Delta = &defDelta
	}

	return m, false
}

func (u *MemStorage) GetAll() []models.Metrics {
	return u.Metrics
}

func (u *MemStorage) AddBatch(metrics []models.Metrics) {

}

func InitMemStorage() *MemStorage {
	var m []models.Metrics
	Storage := &MemStorage{
		Metrics: m,
	}

	if config.GetRestoreMetrics() {
		m, err := GetAllMetrics()

		if err != nil {
			fmt.Println(err)
		}

		Storage.Metrics = m
	}

	return Storage
}
