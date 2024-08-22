package main

import (
	"fmt"

	"github.com/lambawebdev/metrics/internal/agent/services/report"
	"github.com/lambawebdev/metrics/internal/config"
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
