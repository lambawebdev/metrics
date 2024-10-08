package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lambawebdev/metrics/internal/models"
)

const insertGaugeQuery = `
            INSERT INTO metrics (name, type, value) VALUES ($1, $2, $3)
            ON CONFLICT (name)
            DO UPDATE SET value = $3
			`

const insertCounterQuery = `
            INSERT INTO metrics (name, type, delta) VALUES ($1, $2, $3)
            ON CONFLICT (name)
            DO UPDATE SET delta = metrics.delta + $3
			`

type PGSQLMetricRepository struct {
	db *sql.DB
}

func NewPGSQLMetricRepository(db *sql.DB) *PGSQLMetricRepository {
	return &PGSQLMetricRepository{
		db: db,
	}
}

func (repo *PGSQLMetricRepository) AddGauge(metricName string, metricValue float64) {
	var pgErr *pgconn.PgError
	for _, backoff := range backoffSchedule {
		_, err := repo.db.Exec(insertGaugeQuery, metricName, "gauge", metricValue)

		if err != nil {
			fmt.Println(err)
		}

		if err == nil || !errors.As(err, &pgErr) {
			break
		}

		if pgErr.Code == pgerrcode.ConnectionException {
			fmt.Fprintf(os.Stderr, "Request error: %+v\n", err)
			fmt.Fprintf(os.Stderr, "Retrying in %v\n", backoff)
			time.Sleep(backoff)
		}
	}
}

func (repo *PGSQLMetricRepository) AddCounter(metricName string, metricValue int64) {
	var pgErr *pgconn.PgError
	for _, backoff := range backoffSchedule {
		_, err := repo.db.Exec(insertCounterQuery, metricName, "counter", metricValue)

		if err != nil {
			fmt.Println(err)
		}

		if err == nil || !errors.As(err, &pgErr) {
			break
		}

		if pgErr.Code == pgerrcode.ConnectionException {
			fmt.Fprintf(os.Stderr, "Request error: %+v\n", err)
			fmt.Fprintf(os.Stderr, "Retrying in %v\n", backoff)
			time.Sleep(backoff)
		}
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

	return metrics
}

func (repo *PGSQLMetricRepository) GetMetric(metricName string, metricType string) (models.Metrics, bool) {
	var metric models.Metrics
	metric.ID = metricName
	metric.MType = metricType

	defValue := float64(0)
	defDelta := int64(0)

	if metricType == "gauge" {
		metric.Value = &defValue
	}

	if metricType == "counter" {
		metric.Delta = &defDelta
	}

	var pgErr *pgconn.PgError
	for _, backoff := range backoffSchedule {
		row := repo.db.QueryRow("SELECT name, type, delta, value FROM metrics WHERE type = ($1) AND name = ($2)", metricType, metricName)

		err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)

		if err == nil || !errors.As(err, &pgErr) {
			break
		}

		if pgErr.Code == pgerrcode.ConnectionException {
			fmt.Fprintf(os.Stderr, "Request error: %+v\n", err)
			fmt.Fprintf(os.Stderr, "Retrying in %v\n", backoff)
			time.Sleep(backoff)
		}
	}

	return metric, true
}

func (repo *PGSQLMetricRepository) AddBatch(metrics []models.Metrics) {
	tx, err := repo.db.Begin()
	if err != nil {
		fmt.Println(err.Error())
	}

	stmtG, err := tx.Prepare(insertGaugeQuery)
	if err != nil {
		tx.Rollback()
		fmt.Println(err.Error())
	}
	defer stmtG.Close()

	stmtC, err := tx.Prepare(insertCounterQuery)
	if err != nil {
		tx.Rollback()
		fmt.Println(err.Error())
	}
	defer stmtC.Close()

	for _, m := range metrics {
		if m.MType == "gauge" {
			_, err := stmtG.Exec(m.ID, m.MType, m.Value)
			if err != nil {
				tx.Rollback()
				fmt.Println(err.Error())
			}
		}

		if m.MType == "counter" {
			_, err := stmtC.Exec(m.ID, m.MType, m.Delta)
			if err != nil {
				tx.Rollback()
				fmt.Println(err.Error())
			}
		}
	}

	var pgErr *pgconn.PgError
	for _, backoff := range backoffSchedule {
		err = tx.Commit()
		if err != nil {
			fmt.Println(err.Error())
		}

		if err == nil || !errors.As(err, &pgErr) {
			break
		}

		if pgErr.Code == pgerrcode.ConnectionException {
			fmt.Fprintf(os.Stderr, "Request error: %+v\n", err)
			fmt.Fprintf(os.Stderr, "Retrying in %v\n", backoff)
			time.Sleep(backoff)
		}
	}
}

var backoffSchedule = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}
