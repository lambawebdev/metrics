package main

import (
	"net/http"
	"github.com/lambawebdev/metrics/internal/handlers"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`POST /update/{type}/{name}/{value}`, handlers.UpdateMetric)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
