package config

import (
	"errors"
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
	Port          string
	DatabaseURL   string
	LogLevel      string
	Environment   string
	EncryptionKey string
	JWTSecret     string
	NatsURL       string
	FrontendURL   string
	HooksURL      string

	GoogleOAuth    *OAuthProviderConfig
	MicrosoftOAuth *OAuthProviderConfig
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
	HooksURL := os.Getenv("HOOKS_URL")
	if HooksURL == "" {
		HooksURL = "http://localhost:8080"
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/hermes"
	}
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3001"
	}
	log.Printf("Loaded Config: Port=%s", port)
	cfg := &Config{
		Port:          port,
		DatabaseURL:   dbURL,
		LogLevel:      getEnv("LOG_LEVEL", "INFO"),
		Environment:   getEnv("ENV", "development"),
		EncryptionKey: os.Getenv("ENCRYPTION_KEY"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		NatsURL:       natsURL,
		FrontendURL:   frontendURL,
		HooksURL:      HooksURL,
	}
	if id := os.Getenv("GOOGLE_CLIENT_ID"); id != "" {
		cfg.GoogleOAuth = &OAuthProviderConfig{
			ClientID:     id,
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:3000/api/v1/auth/callback/google"),
		}
	}
	if id := os.Getenv("MICROSOFT_CLIENT_ID"); id != "" {
		cfg.MicrosoftOAuth = &OAuthProviderConfig{
			ClientID:     id,
			ClientSecret: os.Getenv("MICROSOFT_CLIENT_SECRET"),
			RedirectURL:  getEnv("MICROSOFT_REDIRECT_URL", "http://localhost:3000/api/v1/auth/callback/microsoft"),
		}
	}
	return cfg
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
	if c.EncryptionKey == "" {
		return errors.New("ENCRYPTION_KEY is required")
	}
	if c.JWTSecret == "" {
		return errors.New("JWT_SECRET is required")
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
