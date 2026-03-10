package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/encryptor"
	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/logger"
	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/oauth"
	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/api"
	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/config"
	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/db"
	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/queue"
	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.LoadConfig()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	appLogger := logger.New("hermes-core", cfg.Environment, cfg.LogLevel)

	appLogger.Info("starting Hermes Core API",
		slog.String("version", "1.0.0"),
		slog.String("port", cfg.Port),
	)
	enc, err := encryptor.NewEncryptor([]byte(cfg.EncryptionKey))
	if err != nil {
		appLogger.Error("encryption init failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	stateCodec, err := oauth.NewStateCodec([]byte(cfg.EncryptionKey))
	if err != nil {
		appLogger.Error("state codec init failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	pool, err := db.New(cfg.DatabaseURL)
	if err != nil {
		appLogger.Error("database connection failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()
	appLogger.Info("database connected")

	providers := make(map[string]oauth.Provider)

	if cfg.GoogleOAuth != nil {
		providers[oauth.ProviderGoogle] = oauth.NewGoogleProvider(oauth.ProviderConfig{
			ClientID:     cfg.GoogleOAuth.ClientID,
			ClientSecret: cfg.GoogleOAuth.ClientSecret,
			RedirectURL:  cfg.GoogleOAuth.RedirectURL,
		})
		appLogger.Info("OAuth provider registered", slog.String("provider", "google"))
	}

	if cfg.MicrosoftOAuth != nil {
		providers[oauth.ProviderMicrosoft] = oauth.NewMicrosoftProvider(oauth.ProviderConfig{
			ClientID:     cfg.MicrosoftOAuth.ClientID,
			ClientSecret: cfg.MicrosoftOAuth.ClientSecret,
			RedirectURL:  cfg.MicrosoftOAuth.RedirectURL,
		})
		appLogger.Info("OAuth provider registered", slog.String("provider", "microsoft"))
	}

	appLogger.Info("OAuth providers loaded", slog.Int("count", len(providers)))

	relayStore := store.NewRelayStore(pool)
	secretStore := store.NewSecretStore(pool, enc)
	userStore := store.NewUserStore(pool)
	connectionStore := store.NewConnectionStore(pool, enc)

	publisher, err := queue.NewNatsPublisher(cfg.NatsURL)
	if err != nil {
		appLogger.Error("failed to create NATS publisher", slog.String("error", err.Error()))
		os.Exit(1)
	}

	handler := api.NewHandler(relayStore, secretStore, userStore, connectionStore, providers, stateCodec, cfg.JWTSecret, appLogger, publisher, cfg.HooksURL, cfg.FrontendURL)
	router := api.NewRouter(handler, cfg.JWTSecret, cfg.FrontendURL)
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		appLogger.Info("server listening", slog.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	appLogger.Info("shutdown signal received")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Error("server shutdown failed", slog.String("error", err.Error()))
	}
	appLogger.Info("server stopped gracefully")
}
