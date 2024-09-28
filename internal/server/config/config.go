package config

import (
	"flag"
	"os"
	"strconv"
)

var options struct {
	flagRunAddr          string
	storeIntervalSeconds uint64
	fileStoragePath      string
	restoreMetrics       bool
	databaseDsn          string
	secretKey            string
}

func ParseFlags() {
	flag.StringVar(&options.flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Uint64Var(&options.storeIntervalSeconds, "i", 300, "save metrics after interval seconds")
	flag.StringVar(&options.fileStoragePath, "f", "/tmp/storage", "file storage path")
	flag.BoolVar(&options.restoreMetrics, "r", true, "if true - metrics will be loaded from file")
	flag.StringVar(&options.databaseDsn, "d", "host=localhost user=test password=password dbname=videos sslmode=disable", "pgsql data source name")
	flag.StringVar(&options.secretKey, "k", "", "set secret key")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		options.flagRunAddr = envRunAddr
	}

	if storeIntervalSeconds := os.Getenv("STORE_INTERVAL"); storeIntervalSeconds != "" {
		value, err := strconv.ParseUint(storeIntervalSeconds, 10, 64)
		if err == nil {
			options.storeIntervalSeconds = value
		}
	}

	if fileStoragePath := os.Getenv("FILE_STORAGE_PATH"); fileStoragePath != "" {
		options.fileStoragePath = fileStoragePath
	}

	if restoreMetrics := os.Getenv("RESTORE"); restoreMetrics != "" {
		value, err := strconv.ParseBool(restoreMetrics)
		if err == nil {
			options.restoreMetrics = value
		}
	}

	if databaseDsn := os.Getenv("DATABASE_DSN"); databaseDsn != "" {
		options.databaseDsn = databaseDsn
	}

	if secretKey := os.Getenv("KEY"); secretKey != "" {
		options.secretKey = secretKey
	}
}

func GetFlagRunAddr() string {
	return options.flagRunAddr
}

func GetStoreIntervalSeconds() uint64 {
	return options.storeIntervalSeconds
}

func GetFileStoragePath() string {
	return options.fileStoragePath
}

func GetRestoreMetrics() bool {
	return options.restoreMetrics
}

func SetRestoreMetrics(bool bool) {
	options.restoreMetrics = bool
}

func GetDatabaseDsn() string {
	return options.databaseDsn
}

func GetSecretKey() string {
	return options.secretKey
}
