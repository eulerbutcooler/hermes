package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/logger"
	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/api"
	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/config"
	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/db"
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

	pool, err := db.New(cfg.DatabaseURL)
	if err != nil {
		appLogger.Error("database connection failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()
	appLogger.Info("database connected")

	relayStore := store.NewRelayStore(pool)
	handler := api.NewHandler(relayStore, appLogger)
	router := api.NewRouter(handler)

	appLogger.Info("server listening", slog.String("port", cfg.Port))
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		appLogger.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
