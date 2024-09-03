package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lambawebdev/metrics/internal/server/config"
	"github.com/lambawebdev/metrics/internal/server/handlers"
	"github.com/lambawebdev/metrics/internal/server/logger"
	"github.com/lambawebdev/metrics/internal/server/middleware"
	"github.com/lambawebdev/metrics/internal/server/storage"
	"go.uber.org/zap"
)

func main() {
	config.ParseFlags()

	r := chi.NewRouter()
	s := storage.InitMemStorage()
	mh := handlers.NewMetricHandler(s)

	go storage.StartToWrite(s, config.GetStoreIntervalSeconds())

	r.Get("/", logger.WithLoggingMiddleware(middleware.GzipMiddleware(func(w http.ResponseWriter, _r *http.Request) {
		mh.GetMetrics(w)
	})))

	r.Post("/value/", logger.WithLoggingMiddleware(middleware.GzipMiddleware(func(w http.ResponseWriter, r *http.Request) {
		mh.GetMetricV2(w, r)
	})))

	r.Get("/value/{type}/{name}", logger.WithLoggingMiddleware(middleware.GzipMiddleware(func(w http.ResponseWriter, r *http.Request) {
		mh.GetMetric(w, r)
	})))

	r.Post("/update/", logger.WithLoggingMiddleware(middleware.GzipMiddleware(func(w http.ResponseWriter, r *http.Request) {
		mh.UpdateMetricV2(w, r)
	})))

	r.Post("/update/{type}/{name}/{value}", logger.WithLoggingMiddleware(middleware.GzipMiddleware(func(w http.ResponseWriter, r *http.Request) {
		mh.UpdateMetric(w, r)
	})))

	err := run(r)
	if err != nil {
		panic(err)
	}
}

func run(handler *chi.Mux) error {
	if err := logger.Initialize("info"); err != nil {
		return err
	}

	logger.Log.Info("Starting server", zap.String("address", config.GetFlagRunAddr()))

	return http.ListenAndServe(config.GetFlagRunAddr(), handler)
}
