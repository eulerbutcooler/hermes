package config

import (
	"errors"
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port        string
	DatabaseURL string
	LogLevel    string
	Environment string
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/hermes"
	}
	log.Printf("Loaded Config: Port=%s", port)
	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
		LogLevel:    getEnv("LOG_LEVEL", "INFO"),
		Environment: getEnv("ENV", "development"),
	}
}

func (c *Config) Validate() error {
	if c.Port == "" {
		return errors.New("PORT can't be empty")
	}
	if _, err := strconv.Atoi(c.Port); err != nil {
		return errors.New(("PORT must be a valid number"))
	}
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL can't be empty")
	}
	validLogLevels := map[string]bool{
		"DEBUG": true,
		"INFO":  true,
		"WARN":  true,
		"ERROR": true,
	}
	if !validLogLevels[c.LogLevel] {
		return errors.New("LOG_LEVEL must be one of: DEBUG, INFO, WARN, ERROR")
	}
	validEnvironments := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
	}
	if !validEnvironments[c.Environment] {
		return errors.New("ENV must be one of: development, staging, production")
	}
	return nil
}
