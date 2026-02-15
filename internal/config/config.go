package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	BaseURL            string
	DatabaseURL        string
	RedisURL           string
	ServerReadTimeout  time.Duration
	ServerWriteTimeout time.Duration
	ServerIdleTimeout  time.Duration
	ShutdownTimeout    time.Duration
	DBTimeout          time.Duration
	CodeLength         int
	MaxGenerateRetries int
}

func Load() (Config, error) {
	godotenv.Load()
	cfg := Config{
		Port:               getEnv("PORT", "8080"),
		BaseURL:            strings.TrimRight(getEnv("BASE_URL", "http://localhost:8080"), "/"),
		DatabaseURL:        strings.TrimSpace(os.Getenv("DATABASE_URL")),
		RedisURL:           strings.TrimSpace(os.Getenv("REDIS_URL")),
		ServerReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
		ServerWriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
		ServerIdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout:    getDurationEnv("SHUTDOWN_TIMEOUT", 10*time.Second),
		DBTimeout:          getDurationEnv("DB_TIMEOUT", 5*time.Second),
		CodeLength:         getIntEnv("CODE_LENGTH", 6),
		MaxGenerateRetries: getIntEnv("MAX_GENERATE_RETRIES", 5),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.CodeLength <= 0 {
		return Config{}, fmt.Errorf("CODE_LENGTH must be > 0")
	}
	if cfg.MaxGenerateRetries <= 0 {
		return Config{}, fmt.Errorf("MAX_GENERATE_RETRIES must be > 0")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
