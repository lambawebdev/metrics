package config

import (
	"flag"
)

var options struct {
	flagRunAddr           string
	pollIntervalSeconds   uint64
	reportIntervalSeconds uint64
}

func ParseFlags() {
	flag.StringVar(&options.flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Uint64Var(&options.pollIntervalSeconds, "p", 2, "poll interval for updating metrics")
	flag.Uint64Var(&options.reportIntervalSeconds, "r", 10, "report interval for sending metrics")

	flag.Parse()
}

func GetFlagRunAddr() string {
	return options.flagRunAddr
}

func GetFlagPollIntervalSeconds() uint64 {
	return options.pollIntervalSeconds
}

func GetFlagReportIntervalSeconds() uint64 {
	return options.reportIntervalSeconds
}
