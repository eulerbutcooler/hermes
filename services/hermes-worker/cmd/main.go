package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/encryptor"
	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/logger"
	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/oauth"
	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/config"
	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/engine"
	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/integrations/debug"
	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/integrations/discord"
	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/integrations/email"
	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/integrations/httpreq"
	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/integrations/slack"
	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/queue"
	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.LoadConfig()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	appLogger := logger.New("hermes-worker", cfg.Environment, cfg.LogLevel)
	appLogger.Info("starting Hermes Worker",
		slog.String("version", "1.0.0"),
		slog.String("environment", cfg.Environment),
	)

	enc, err := encryptor.NewEncryptor([]byte(cfg.EncryptionKey))
	if err != nil {
		appLogger.Error("encryption init failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	db, err := store.NewStore(cfg.DbURL, enc)
	if err != nil {
		appLogger.Error("database initialization failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
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

	//Registry Pattern
	// Registering integrations instead of hardcoding
	reg := engine.NewRegistry()
	reg.Register("debug_log", debug.New())
	reg.Register("discord_send", discord.New())
	reg.Register("slack_send", slack.New())
	reg.Register("http_request", httpreq.New())
	reg.Register("email_send", email.New(providers, db))
	appLogger.Info("integrations loaded",
		slog.Int("count", reg.Count()),
		slog.Any("types", reg.Types()),
	)

	pool := engine.NewWorkerPool(10, db, reg, appLogger)
	ctx, cancel := context.WithCancel(context.Background())
	pool.Start(ctx)

	cronScheduler := engine.NewCronScheduler(db, pool.JobQueue, appLogger)
	cronScheduler.Start(ctx)

	consumer, err := queue.NewConsumer(cfg.NatsURL, pool.JobQueue, appLogger)
	if err != nil {
		appLogger.Error("NATS consumer creation failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := consumer.Start(); err != nil {
		appLogger.Error("failed to start consumer", slog.String("error", err.Error()))
		os.Exit(1)
	}
	appLogger.Info("Hermes Worker is running", slog.String("status", "ready"))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	appLogger.Info("shutdown signal received, initiating graceful shutdown")
	if err := consumer.Stop(); err != nil {
		appLogger.Error("error stopping consumer", slog.String("error", err.Error()))
	}
	cronScheduler.Stop()
	appLogger.Info("producers stopped, draining queue")
	pool.Shutdown()
	cancel()
	appLogger.Info("Worker stoppped gracefully")
}
