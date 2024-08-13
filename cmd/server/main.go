package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lambawebdev/metrics/internal/config"
	"github.com/lambawebdev/metrics/internal/handlers"
	"github.com/lambawebdev/metrics/internal/storage"
)

func main() {
	config.ParseFlags()

	r := chi.NewRouter()

	storage := new(storage.MemStorage)
	storage.Metrics = make(map[string]interface{})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetMetrics(w, r, storage)
	})

	r.Get("/value/{type}/{name}", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetMetric(w, r, storage)
	})

	r.Post("/update/{type}/{name}/{value}", func(w http.ResponseWriter, r *http.Request) {
		handlers.UpdateMetric(w, r, storage)
	})

	err := http.ListenAndServe(config.GetFlagRunAddr(), r)
	if err != nil {
		panic(err)
	}
}
