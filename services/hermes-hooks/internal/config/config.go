package config

import "os"

type Config struct {
	Port        string
	NatsUrl     string
	Environment string
	LogLevel    string
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Loads and Validates env variables
func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	natsUrl := os.Getenv("NATS_URL")
	if natsUrl == "" {
		natsUrl = "nats://localhost:4222"
	}
	return &Config{
		Port:        port,
		NatsUrl:     natsUrl,
		Environment: getEnv("ENV", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "INFO"),
	}
}
