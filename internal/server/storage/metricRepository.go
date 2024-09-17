package storage

import (
	"database/sql"
	"fmt"

	"github.com/lambawebdev/metrics/internal/models"
)

type PGSQLMetricRepository struct {
	db *sql.DB
}

func NewPGSQLMetricRepository(db *sql.DB) *PGSQLMetricRepository {
	return &PGSQLMetricRepository{
		db: db,
	}
}

func (repo *PGSQLMetricRepository) AddGauge(metricName string, metricValue float64) {
	fmt.Println("GAUGE", metricName, metricValue)
	_, err := repo.db.Exec(`
	        INSERT INTO metrics (name, type, value) VALUES ($1, $2, $3)
            ON CONFLICT (name)
            DO UPDATE SET value = $3
		`, metricName, "gauge", metricValue)

	if err != nil {
		fmt.Println(err)
	}
}

func (repo *PGSQLMetricRepository) AddCounter(metricName string, metricValue int64) {
	fmt.Println("COUNTER", metricName, metricValue)

	_, err := repo.db.Exec(`
	        INSERT INTO metrics (name, type, delta) VALUES ($1, $2, $3)
            ON CONFLICT (name)
            DO UPDATE SET delta = metrics.delta + $3
		`, metricName, "counter", metricValue)

	if err != nil {
		fmt.Println(err)
	}
}

func (repo *PGSQLMetricRepository) GetAll() []models.Metrics {
	rows, err := repo.db.Query("SELECT name, type, delta, value FROM metrics")

	if err != nil {
		fmt.Println(err)
	}

	defer rows.Close()

	var metrics []models.Metrics

	for rows.Next() {
		var metric models.Metrics
		if err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
			fmt.Println(err)
		}

		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		fmt.Println(err)
	}

	fmt.Println("GET ALL!!!", metrics)

	return metrics
}

func (repo *PGSQLMetricRepository) GetMetric(metricName string, metricType string) (models.Metrics, bool) {
	var metric models.Metrics

	row := repo.db.QueryRow("SELECT name, type, delta, value FROM metrics WHERE type = ($1) AND name = ($2)", metricType, metricName)

	if err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
		fmt.Println(err)
	}

	return metric, true
}
