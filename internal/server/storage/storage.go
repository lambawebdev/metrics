package storage

import (
	"fmt"
	"reflect"

	"github.com/lambawebdev/metrics/internal/server/config"
)

type MetricWriter interface {
	AddGauge(metricName string, metricValue float64)
	AddCounter(metricName string, metricValue int64)
	GetMetricValue(metricName string)
	GetAll()
}

type MemStorage struct {
	Metrics map[string]interface{}
}

func (u *MemStorage) AddGauge(metricName string, metricValue float64) {
	u.Metrics[metricName] = metricValue
}

func (u *MemStorage) AddCounter(metricName string, metricValue int64) {
	if u.Metrics[metricName] != nil {
		if reflect.TypeOf(u.Metrics[metricName]).Kind() == reflect.Float64 {
			metricValue = int64(u.Metrics[metricName].(float64)) + metricValue
		}

		if reflect.TypeOf(u.Metrics[metricName]).Kind() == reflect.Int64 {
			metricValue = u.Metrics[metricName].(int64) + metricValue
		}
	}

	u.Metrics[metricName] = metricValue
}

func (u *MemStorage) GetMetricValue(metricName string) interface{} {
	return u.Metrics[metricName]
}

func (u *MemStorage) GetAll() map[string]interface{} {
	return u.Metrics
}

func InitMemStorage() *MemStorage {
	Storage := new(MemStorage)
	Storage.Metrics = make(map[string]interface{})

	if config.GetRestoreMetrics() {
		m, err := GetAllMetrics()

		if err != nil {
			fmt.Println(err)
		}

		Storage.Metrics = m
	}

	return Storage
}
