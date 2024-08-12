package storage

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
		metricValue = u.Metrics[metricName].(int64) + metricValue
	}

	u.Metrics[metricName] = metricValue
}

func (u *MemStorage) GetMetricValue(metricName string) interface{} {
	return u.Metrics[metricName]
}

func (u *MemStorage) GetAll() map[string]interface{} {
	return u.Metrics
}
