package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type Config struct {
	Environment  string
	NatsURL      string
	DbURL        string
	MaxWorkers   int
	JobQueueSize int
	LogLevel     string
	LogPretty    bool
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if val := os.Getenv(key); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func LoadConfig() *Config {
	cfg := &Config{
		Environment:  getEnv("ENV", "development"),
		NatsURL:      getEnv("NATS_URL", "nats://localhost:4222"),
		DbURL:        getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/hermes"),
		MaxWorkers:   getEnvInt("MAX_WORKERS", 10),
		JobQueueSize: getEnvInt("JOB_QUEUE_SIZE", 100),
		LogLevel:     getEnv("LOG_LEVEL", "INFO"),
	}
	log.Printf("Loaded Config: Environment: %s, MaxWorkers: %d", cfg.Environment, cfg.MaxWorkers)
	return cfg
}

func (c *Config) Validate() error {
	if c.DbURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.NatsURL == "" {
		return fmt.Errorf("NATS_URL is required")
	}
	if c.MaxWorkers < 1 {
		return fmt.Errorf("MAX_WORKERS must be atleast 1")
	}
	return nil
}
