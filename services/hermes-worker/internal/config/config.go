package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type Config struct {
	Environment   string
	EncryptionKey string
	NatsURL       string
	DbURL         string
	MaxWorkers    int
	JobQueueSize  int
	LogLevel      string
	LogPretty     bool

	GoogleOAuth    *OAuthProviderConfig
	MicrosoftOAuth *OAuthProviderConfig
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
		Environment:   getEnv("ENV", "development"),
		EncryptionKey: os.Getenv("ENCRYPTION_KEY"),
		NatsURL:       getEnv("NATS_URL", "nats://localhost:4222"),
		DbURL:         getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/hermes"),
		MaxWorkers:    getEnvInt("MAX_WORKERS", 10),
		JobQueueSize:  getEnvInt("JOB_QUEUE_SIZE", 100),
		LogLevel:      getEnv("LOG_LEVEL", "INFO"),
	}

	if id := os.Getenv("GOOGLE_CLIENT_ID"); id != "" {
		cfg.GoogleOAuth = &OAuthProviderConfig{
			ClientID:     id,
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8081/api/v1/auth/callback/google"),
		}
	}
	if id := os.Getenv("MICROSOFT_CLIENT_ID"); id != "" {
		cfg.MicrosoftOAuth = &OAuthProviderConfig{
			ClientID:     id,
			ClientSecret: os.Getenv("MICROSOFT_CLIENT_SECRET"),
			RedirectURL:  getEnv("MICROSOFT_REDIRECT_URL", "http://localhost:8081/api/v1/auth/callback/microsoft"),
		}
	}

	log.Printf("Loaded Config: Environment: %s, MaxWorkers: %d", cfg.Environment, cfg.MaxWorkers)
	return cfg
}

func (c *Config) Validate() error {
	if c.EncryptionKey == "" {
		return fmt.Errorf("ENCRYPTION_KEY is required")
	}
	if c.NatsURL == "" {
		return fmt.Errorf("NATS_URL is required")
	}
	if c.DbURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.MaxWorkers < 1 {
		return fmt.Errorf("MAX_WORKERS must be atleast 1")
	}
	return nil
}
