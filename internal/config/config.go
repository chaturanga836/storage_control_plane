package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddress         string
	SharedRocksDBPath     string
	SharedClickHouseDSN   string
	PostgresDSN           string
	
	// Additional configuration
	Environment           string
	LogLevel              string
	WALFlushThreshold     int
	ParquetBatchSize      int
}

func Load() (*Config, error) {
	// Load .env file if it exists (optional - won't fail if missing)
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading it: %v (this is optional)", err)
	}

	return &Config{
		ServerAddress:       getEnv("SERVER_ADDR", ":8081"),
		SharedRocksDBPath:   getEnv("SHARED_ROCKSDB_PATH", "/data/shared_rocksdb"),
		SharedClickHouseDSN: getEnv("SHARED_CLICKHOUSE_DSN", "tcp://localhost:9000?debug=true"),
		PostgresDSN:         getEnv("POSTGRES_DSN", "postgres://user:pass@localhost:5432/core"),
		
		// Additional settings
		Environment:         getEnv("GO_ENV", "development"),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		WALFlushThreshold:  getEnvInt("WAL_FLUSH_THRESHOLD", 1000),
		ParquetBatchSize:   getEnvInt("PARQUET_BATCH_SIZE", 10000),
	}, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
		log.Printf("Invalid integer value for %s: %s, using default %d", key, v, def)
	}
	return def
}