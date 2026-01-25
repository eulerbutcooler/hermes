package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/logger"
	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/config"
	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/engine"
	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/integrations/debug"
	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/integrations/discord"
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

	db, err := store.NewStore(cfg.DbURL)
	if err != nil {
		appLogger.Error("database initialization failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	appLogger.Info("database connected")

	//Registry Pattern
	// Registering integrations instead of hardcoding
	reg := engine.NewRegistry()
	reg.Register("debug_log", debug.New())
	reg.Register("discord_send", discord.New())
	reg.Register("slack_send", slack.New())
	appLogger.Info("integrations loaded",
		slog.Int("count", 3),
		slog.Any("types", []string{"debug_log", "discord_send", "slack_send"}),
	)

	pool := engine.NewWorkerPool(10, db, reg, appLogger)
	ctx, cancel := context.WithCancel(context.Background())
	pool.Start(ctx)

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
	cancel()
	pool.Shutdown()
	appLogger.Info("Worker stoppped gracefully")
}
