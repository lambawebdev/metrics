package main

import (
	"fmt"

	"github.com/lambawebdev/metrics/internal/agent/config"
	"github.com/lambawebdev/metrics/internal/agent/services/report"
)

func main() {
	config.ParseFlags()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()

	report.Start()
}
