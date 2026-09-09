package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	PGHost     string
	PGPort     string
	PGUser     string
	PGPassword string
	PGDatabase string
	Port       string
	LogLevel   *slog.LevelVar
}

func Load() *Config {
	return &Config{
		Port:       getEnv("PORT", "8080"),
		LogLevel:   parseLogLevel(),
		PGHost:     getEnv("PGHOST", "localhost"),
		PGPort:     getEnv("PGPORT", "5432"),
		PGUser:     getEnv("PGUSER", "gtfs"),
		PGPassword: getEnv("PGPASSWORD", "gtfs"),
		PGDatabase: getEnv("PGDATABASE", "gtfs"),
	}
}

func parseLogLevel() (logLevel *slog.LevelVar) {
	logLevelStr := getEnv("LOG_LEVEL", "info")
	level := new(slog.LevelVar)

	switch strings.ToLower(logLevelStr) {
	case "debug":
		level.Set(slog.LevelDebug)
	case "info":
		level.Set(slog.LevelInfo)
	case "warn", "warning":
		level.Set(slog.LevelWarn)
	case "error":
		level.Set(slog.LevelError)
	default:
		level.Set(slog.LevelInfo)
	}

	return level
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.PGUser, c.PGPassword, c.PGHost, c.PGPort, c.PGDatabase)
}
