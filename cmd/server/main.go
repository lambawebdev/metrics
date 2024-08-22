package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lambawebdev/metrics/internal/config"
	"github.com/lambawebdev/metrics/internal/server/handlers"
	"github.com/lambawebdev/metrics/internal/server/logger"
	"github.com/lambawebdev/metrics/internal/server/storage"
	"go.uber.org/zap"
)

func main() {
	r := chi.NewRouter()

	storage := new(storage.MemStorage)
	storage.Metrics = make(map[string]interface{})

	r.Get("/", logger.WithLoggingMiddleware(func(w http.ResponseWriter, _r *http.Request) {
		handlers.GetMetrics(w, storage)
	}))

	r.Post("/value/", logger.WithLoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlers.GetMetricV2(w, r, storage)
	}))

	r.Get("/value/{type}/{name}", logger.WithLoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlers.GetMetric(w, r, storage)
	}))

	r.Post("/update/", logger.WithLoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlers.UpdateMetricV2(w, r, storage)
	}))

	r.Post("/update/{type}/{name}/{value}", logger.WithLoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlers.UpdateMetric(w, r, storage)
	}))

	err := run(r)
	if err != nil {
		panic(err)
	}
}

func run(handler *chi.Mux) error {
	config.ParseFlags()

	if err := logger.Initialize("info"); err != nil {
		return err
	}

	logger.Log.Info("Starting server", zap.String("address", config.GetFlagRunAddr()))

	return http.ListenAndServe(config.GetFlagRunAddr(), handler)
}
