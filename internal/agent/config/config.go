package config

import (
	"flag"
	"os"
	"strconv"
)

var options struct {
	flagRunAddr           string
	pollIntervalSeconds   uint64
	reportIntervalSeconds uint64
	secretKey             string
}

func ParseFlags() {
	flag.StringVar(&options.flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Uint64Var(&options.pollIntervalSeconds, "p", 2, "poll interval for updating metrics")
	flag.Uint64Var(&options.reportIntervalSeconds, "r", 10, "report interval for sending metrics")
	flag.StringVar(&options.secretKey, "k", "", "set secret key")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		options.flagRunAddr = envRunAddr
	}

	if envPollIntervalSeconds := os.Getenv("POLL_INTERVAL"); envPollIntervalSeconds != "" {
		value, err := strconv.ParseUint(envPollIntervalSeconds, 10, 64)
		if err == nil {
			options.pollIntervalSeconds = value
		}
	}

	if envReportIntervalSeconds := os.Getenv("REPORT_INTERVAL"); envReportIntervalSeconds != "" {
		value, err := strconv.ParseUint(envReportIntervalSeconds, 10, 64)
		if err == nil {
			options.reportIntervalSeconds = value
		}
	}

	if secretKey := os.Getenv("KEY"); secretKey != "" {
		options.secretKey = secretKey
	}
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

func GetSecretKey() string {
	return options.secretKey
}
