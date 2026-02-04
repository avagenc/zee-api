package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func LoadPGPool() (*pgxpool.Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	cfg.MaxConns = getEnvInt32("DATABASE_MAX_CONNS", 20)
	cfg.MinConns = getEnvInt32("DATABASE_MIN_CONNS", 0)
	cfg.MaxConnLifetime = getEnvDuration("DATABASE_MAX_CONN_LIFETIME", 1*time.Hour)
	cfg.MaxConnIdleTime = getEnvDuration("DATABASE_MAX_CONN_IDLE_TIME", 30*time.Minute)

	return cfg, nil
}

func getEnvInt32(key string, defaultValue int32) int32 {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}

	intVal, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return defaultValue
	}

	return int32(intVal)
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}

	duration, err := time.ParseDuration(val)
	if err != nil {
		return defaultValue
	}

	return duration
}
