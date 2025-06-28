package config

import (
	"os"
)

type Config struct {
	ServerAddress         string
	SharedRocksDBPath     string
	SharedClickHouseDSN   string
	PostgresDSN           string
	// Add more config as needed
}

func Load() (*Config, error) {
	return &Config{
		ServerAddress:       getEnv("SERVER_ADDR", ":8080"),
		SharedRocksDBPath:   getEnv("SHARED_ROCKSDB_PATH", "/data/shared_rocksdb"),
		SharedClickHouseDSN: getEnv("SHARED_CLICKHOUSE_DSN", "tcp://localhost:9000?debug=true"),
		PostgresDSN:         getEnv("POSTGRES_DSN", "postgres://user:pass@localhost:5432/core"),
	}, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}