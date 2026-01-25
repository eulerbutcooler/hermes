package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/models"
	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/store"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	store   *store.RelayStore
	logger  *slog.Logger
	baseURL string
}

func NewHandler(s *store.RelayStore, logger *slog.Logger) *Handler {
	return &Handler{store: s, logger: logger, baseURL: "http://localhost:8080"}
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", slog.String("error", err.Error()))
	}
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message, code string) {
	h.respondJSON(w, status, models.ErrorResponse{
		Success: false,
		Error:   message,
		Code:    code,
	})
}

func (h *Handler) respondSuccess(w http.ResponseWriter, status int, message string, data any) {
	h.respondJSON(w, status, models.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
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
	if strings.TrimSpace(req.Name) == "" {
		h.respondError(w, http.StatusBadRequest, "Name is required", "VALIDATION_ERROR")
		return
	}
	if strings.TrimSpace(req.UserID) == "" {
		h.respondError(w, http.StatusBadRequest, "UserID is required", "VALIDATION_ERROR")
		return
	}
	if len(req.Actions) == 0 {
		h.respondError(w, http.StatusBadRequest, "At least one action is required", "VALIDATION_ERROR")
		return
	}

	for i, action := range req.Actions {
		if action.ActionType == "" {
			h.respondError(w, http.StatusBadRequest,
				"Action type is required for action at index "+strconv.Itoa(i),
				"VALIDATION_ERROR")
			return
		}
		if action.Config == nil {
			h.respondError(w, http.StatusBadRequest,
				"Config is required for action at index "+strconv.Itoa(i),
				"VALIDATION_ERROR")
			return
		}
	}

	relay, err := h.store.CreateRelay(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to create relay",
			slog.String("error", err.Error()),
			slog.String("user_id", req.UserID),
		)
		h.respondError(w, http.StatusInternalServerError, "Failed to create relay", "DB_ERROR")
		return
	}
	relay.Relay.WebhookURL = h.baseURL + relay.Relay.WebhookPath

	h.logger.Info("relay created",
		slog.String("relay_id", relay.ID),
		slog.String("user_id", req.UserID),
		slog.Int("action_count", len(relay.Actions)),
	)

	h.respondSuccess(w, http.StatusCreated, "Relay created successfully", relay)

}

func (h *Handler) GetAllRelays(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")

	h.logger.Debug("fetching all relays",
		slog.String("user_id", userID),
	)

	relays, err := h.store.GetAllRelays(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to fetch relays",
			slog.String("error", err.Error()),
		)
		h.respondError(w, http.StatusInternalServerError, "Failed to fetch relays", "DB_ERROR")
		return
	}

	for i := range relays {
		relays[i].WebhookURL = h.baseURL + relays[i].WebhookPath
	}

	h.logger.Info("fetched relays",
		slog.Int("count", len(relays)),
		slog.String("user_id", userID),
	)

	h.respondSuccess(w, http.StatusOK, "", relays)
}

func (h *Handler) GetRelayLogs(w http.ResponseWriter, r *http.Request) {
	relayID := chi.URLParam(r, "id")
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = min(parsedLimit, 200)
		}
	}
	h.logger.Debug("fetching relay logs", slog.String("relay_id", relayID),
		slog.Int("limit", limit))
	logs, err := h.store.GetLogs(r.Context(), relayID, limit)
	if err != nil {
		h.logger.Error("failed to fetch logs", slog.String("relay_id", relayID),
			slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to fetch logs", "DB_ERROR")
		return
	}
	h.logger.Info("fetched logs", slog.String("relay_id", relayID), slog.Int("count", len(logs)))
	h.respondSuccess(w, http.StatusOK, "", logs)
}

func (h *Handler) GetRelay(w http.ResponseWriter, r *http.Request) {
	relayID := chi.URLParam(r, "id")
	h.logger.Debug("fetching relay", slog.String("relay_id", relayID))
	relay, err := h.store.GetRelay(r.Context(), relayID)
	if err != nil {
		if errors.Is(err, store.ErrRelayNotFound) {
			h.logger.Warn("relay not found", slog.String("relay_id", relayID))
			h.respondError(w, http.StatusNotFound, "Relay Not found", "NOT_FOUND")
			return
		}
		h.logger.Error("failed to fetch relay",
			slog.String("relay_id", relayID),
			slog.String("error", err.Error()),
		)
		h.respondError(w, http.StatusInternalServerError, "Failed to fetch relay", "DB_ERROR")
		return
	}
	relay.Relay.WebhookURL = h.baseURL + relay.Relay.WebhookPath
	h.logger.Info("fetched relay",
		slog.String("relay_id", relayID),
		slog.Int("action_count", len(relay.Actions)),
	)

	h.respondSuccess(w, http.StatusOK, "", relay)
}

func (h *Handler) UpdateRelay(w http.ResponseWriter, r *http.Request) {
	relayID := chi.URLParam(r, "id")
	var req models.UpdateRelayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid request body", slog.String("error", err.Error()))
		h.respondError(w, http.StatusBadRequest, "Invalid JSON body", "INVALID_JSON")
		return
	}
	if req.Name == nil && req.Description == nil && req.IsActive == nil {
		h.respondError(w, http.StatusBadRequest, "No fields to update", "VALIDATION_ERROR")
		return
	}
	relay, err := h.store.UpdateRelay(r.Context(), relayID, req)
	if err != nil {
		if errors.Is(err, store.ErrRelayNotFound) {
			h.logger.Warn("relay not found", slog.String("relay_id", relayID))
			h.respondError(w, http.StatusNotFound, "Relay not found", "NOT_FOUND")
			return
		}
		h.logger.Error("failed to update relay", slog.String("relay_id", relayID),
			slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to update relay", "DB_ERROR")
		return
	}
	relay.WebhookURL = h.baseURL + relay.WebhookPath
	h.logger.Info("relay updated", slog.String("relay_id", relayID))
	h.respondSuccess(w, http.StatusOK, "Relay updated successfully", relay)
}

func (h *Handler) DeleteRelay(w http.ResponseWriter, r *http.Request) {
	relayID := chi.URLParam(r, "id")
	err := h.store.DeleteRelay(r.Context(), relayID)
	if err != nil {
		if errors.Is(err, store.ErrRelayNotFound) {
			h.logger.Warn("relay not found for deletion", slog.String("relay_id", relayID))
			h.respondError(w, http.StatusNotFound, "Relay not found", "NOT_FOUND")
			return
		}
		h.logger.Error("failed to delete relay", slog.String("relay_id", relayID),
			slog.String("err", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to delete relay", "DB_ERROR")
		return
	}
	h.logger.Info("relay deleted", slog.String("relay_id", relayID))
	h.respondSuccess(w, http.StatusOK, "Relay deleted successfully",
		map[string]string{
			"deleted_id": relayID,
		})
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "hermes-core",
	})
}
