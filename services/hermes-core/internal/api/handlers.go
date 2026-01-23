package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/models"
	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/store"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	store  *store.RelayStore
	logger *slog.Logger
}

func NewHandler(s *store.RelayStore, logger *slog.Logger) *Handler {
	return &Handler{store: s, logger: logger}
}

func (h *Handler) CreateRelay(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRelayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid request body",
			slog.String("error", err.Error()),
			slog.String("path", r.URL.Path),
		)
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.ActionType == "" {
		h.logger.Warn("missing required fields",
			slog.String("name", req.Name),
			slog.String("action_type", req.ActionType),
		)
		http.Error(w, "Name and ActionType are required", http.StatusBadRequest)
		return
	}
	id, err := h.store.CreateRelay(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to create relay",
			slog.String("error", err.Error()),
			slog.Int("user_id", req.UserID),
		)
		http.Error(w, "Failed to create relay: "+err.Error(), http.StatusInternalServerError)
	}

	h.logger.Info("relay created",
		slog.String("relay_id", id),
		slog.Int("user_id", req.UserID),
		slog.String("action_type", req.ActionType),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "success",
		"relay_id": id,
	})
}

func (h *Handler) GetRelayLogs(w http.ResponseWriter, r *http.Request) {
	relayID := chi.URLParam(r, "relayID")
	h.logger.Debug("fetching relay logs",
		slog.String("relay_id", relayID),
	)
	logs, err := h.store.GetLogs(r.Context(), relayID)
	if err != nil {
		h.logger.Error("failed to fetch logs",
			slog.String("relay_id", relayID),
			slog.String("error", err.Error()),
		)
		http.Error(w, fmt.Sprintf("Failed to fetch logs: %v", err), http.StatusInternalServerError)
		return
	}
	h.logger.Info("fetched logs successfully",
		slog.String("relay_id", relayID),
		slog.Int("count", len(logs)),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func (*Handler) GetRelay(w *http.ResponseWriter, r *http.Request)
