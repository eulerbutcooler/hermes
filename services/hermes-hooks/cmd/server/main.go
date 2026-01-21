package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/logger"
	"github.com/eulerbutcooler/hermes/services/hermes-hooks/internal/api"
	"github.com/eulerbutcooler/hermes/services/hermes-hooks/internal/config"
	"github.com/eulerbutcooler/hermes/services/hermes-hooks/internal/queue"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.LoadConfig()
	appLogger := logger.New("hermes-hooks", cfg.Environment, cfg.LogLevel)

	appLogger.Info("starting Hermes Hooks",
		slog.String("version", "1.0.0"),
		slog.String("port", cfg.Port),
	)

	natsQueue, err := queue.NewNatsQueue(cfg.NatsUrl)
	if err != nil {
		appLogger.Error("NATS connection failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	appLogger.Info("connected to NATS", slog.String("url", cfg.NatsUrl))

	handler := api.NewHandler(natsQueue, appLogger)
	r := api.NewRouter(handler)

	appLogger.Info("webhook server listening", slog.String("port", cfg.Port))
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		appLogger.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
