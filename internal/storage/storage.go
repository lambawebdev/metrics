package storage

type MetricWriter interface {
	AddGauge(metricName string, metricValue float64)
	AddCounter(metricName string, metricValue int64)
}

type MemStorage struct {
	GaugeMetric   map[string]float64
	CounterMetric map[string]int64
}

func (u *MemStorage) AddGauge(metricName string, metricValue float64) {
	u.GaugeMetric[metricName] = metricValue
}

func (u *MemStorage) AddCounter(metricName string, metricValue int64) {
	u.CounterMetric[metricName] += metricValue
}
