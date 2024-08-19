package main

import (
	"github.com/lambawebdev/metrics/internal/config"
	"github.com/lambawebdev/metrics/internal/services/report"
)

func main() {
	config.ParseFlags()

	var m report.Monitor

	go report.NewMonitor(&m)
	report.Report(&m)
}
