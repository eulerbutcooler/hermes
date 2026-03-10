package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/actions"
	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/models"
	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/store"
	"github.com/go-chi/chi/v5"
)

//  RESPONSE HELPERS

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

// RELAY API

func (h *Handler) CreateRelay(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRelayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid request body",
			slog.String("error", err.Error()),
			slog.String("path", r.URL.Path),
		)
		h.respondError(w, http.StatusBadRequest, "Invalid JSON body", "INVALID_JSON")
		return
	}

	userID := GetUserID(r)
	req.UserID = userID

	if strings.TrimSpace(req.Name) == "" {
		h.respondError(w, http.StatusBadRequest, "Name is required", "VALIDATION_ERROR")
		return
	}
	if len(req.Actions) == 0 {
		h.respondError(w, http.StatusBadRequest, "At least one action is required", "VALIDATION_ERROR")
		return
	}

	switch req.TriggerType {
	case "", models.TriggerWebhook, models.TriggerManual:
	case models.TriggerCron:
		schedule, ok := req.TriggerConfig["schedule"].(string)
		if !ok || schedule == "" {
			h.respondError(w, http.StatusBadRequest,
				"trigger_config.schedule is required for cron triggers", "VALIDATION_ERROR")
			return
		}
	default:
		h.respondError(w, http.StatusBadRequest,
			"unknown trigger_type: "+string(req.TriggerType), "VALIDATION_ERROR")
		return
	}

	for i, action := range req.Actions {
		if action.ActionType == "" {
			h.respondError(w, http.StatusBadRequest,
				"Action type is required for action at index "+strconv.Itoa(i),
				"VALIDATION_ERROR")
			return
		}
		if !actions.IsValidType(action.ActionType) {
			h.respondError(w, http.StatusBadRequest,
				"Unknown action type '"+action.ActionType+"' at index "+strconv.Itoa(i),
				"VALIDATION_ERROR")
		}
		if action.Config == nil {
			h.respondError(w, http.StatusBadRequest,
				"Config is required for action at index "+strconv.Itoa(i),
				"VALIDATION_ERROR")
			return
		}
		if err := actions.ValidateConfig(action.ActionType, action.Config); err != nil {
			h.respondError(w, http.StatusBadRequest,
				"Invalid config for action at index"+strconv.Itoa(i)+": "+err.Error(),
				"VALIDATION_ERROR")
			return
		}
	}

	relay, err := h.store.CreateRelay(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to create relay",
			slog.String("error", err.Error()),
			slog.String("user_id", userID),
		)
		h.respondError(w, http.StatusInternalServerError, "Failed to create relay", "DB_ERROR")
		return
	}
	relay.Relay.WebhookURL = h.hooksURL + relay.Relay.WebhookPath

	h.logger.Info("relay created",
		slog.String("relay_id", relay.ID),
		slog.String("user_id", userID),
		slog.Int("action_count", len(relay.Actions)),
	)

	h.respondSuccess(w, http.StatusCreated, "Relay created successfully", relay)

}

func (h *Handler) GetAllRelays(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)

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
		relays[i].WebhookURL = h.hooksURL + relays[i].WebhookPath
	}

	h.logger.Info("fetched relays",
		slog.Int("count", len(relays)),
		slog.String("user_id", userID),
	)

	h.respondSuccess(w, http.StatusOK, "", relays)
}

func (h *Handler) GetExecutions(w http.ResponseWriter, r *http.Request) {
	relayID := chi.URLParam(r, "id")
	userID := GetUserID(r)

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = min(parsedLimit, 200)
		}
	}

	h.logger.Debug("fetching executions",
		slog.String("relay_id", relayID),
		slog.Int("limit", limit),
	)

	executions, err := h.store.GetExecutions(r.Context(), relayID, userID, limit)
	if err != nil {
		h.logger.Error("failed to fetch executions",
			slog.String("relay_id", relayID),
			slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to fetch executions", "DB_ERROR")
		return
	}

	h.respondSuccess(w, http.StatusOK, "", executions)
}

func (h *Handler) GetExecutionSteps(w http.ResponseWriter, r *http.Request) {
	executionID := chi.URLParam(r, "executionId")
	userID := GetUserID(r)

	h.logger.Debug("fetching execution steps",
		slog.String("execution_id", executionID),
	)

	steps, err := h.store.GetExecutionSteps(r.Context(), executionID, userID)
	if err != nil {
		h.logger.Error("failed to fetch execution steps",
			slog.String("execution_id", executionID),
			slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to fetch execution steps", "DB_ERROR")
		return
	}

	h.respondSuccess(w, http.StatusOK, "", steps)
}

func (h *Handler) DeleteExecution(w http.ResponseWriter, r *http.Request) {
	executionID := chi.URLParam(r, "executionId")
	userID := GetUserID(r)

	err := h.store.DeleteExecution(r.Context(), executionID, userID)
	if err != nil {
		if errors.Is(err, store.ErrExecutionNotFound) {
			h.respondError(w, http.StatusNotFound, "Execution not found", "NOT_FOUND")
			return
		}
		h.logger.Error("failed to delete execution",
			slog.String("execution_id", executionID),
			slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to delete execution", "DB_ERROR")
		return
	}
	h.respondSuccess(w, http.StatusOK, "Execution deleted", map[string]string{"deleted_id": executionID})
}

func (h *Handler) GetRelay(w http.ResponseWriter, r *http.Request) {
	relayID := chi.URLParam(r, "id")
	userID := GetUserID(r)
	h.logger.Debug("fetching relay", slog.String("relay_id", relayID))
	relay, err := h.store.GetRelay(r.Context(), relayID, userID)
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
	relay.Relay.WebhookURL = h.hooksURL + relay.Relay.WebhookPath
	h.logger.Info("fetched relay",
		slog.String("relay_id", relayID),
		slog.Int("action_count", len(relay.Actions)),
	)

	h.respondSuccess(w, http.StatusOK, "", relay)
}

func (h *Handler) UpdateRelay(w http.ResponseWriter, r *http.Request) {
	relayID := chi.URLParam(r, "id")
	userID := GetUserID(r)
	var req models.UpdateRelayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid request body", slog.String("error", err.Error()))
		h.respondError(w, http.StatusBadRequest, "Invalid JSON body", "INVALID_JSON")
		return
	}
	if req.Name == nil && req.Description == nil && req.IsActive == nil && req.TriggerType == nil {
		h.respondError(w, http.StatusBadRequest, "No fields to update", "VALIDATION_ERROR")
		return
	}

	if req.TriggerType != nil {
		switch *req.TriggerType {
		case models.TriggerWebhook, models.TriggerManual:
		case models.TriggerCron:
			schedule, _ := req.TriggerConfig["schedule"].(string)
			if schedule == "" {
				h.respondError(w, http.StatusBadRequest,
					"trigger_config.schedule is required for cron triggers", "VALIDATION_ERROR")
				return
			}
		default:
			h.respondError(w, http.StatusBadRequest,
				"unknown trigger_type: "+string(*req.TriggerType), "VALIDATION_ERROR")
			return
		}
	}

	relay, err := h.store.UpdateRelay(r.Context(), relayID, userID, req)
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
	relay.WebhookURL = h.hooksURL + relay.WebhookPath
	h.logger.Info("relay updated", slog.String("relay_id", relayID))
	h.respondSuccess(w, http.StatusOK, "Relay updated successfully", relay)
}

func (h *Handler) UpdateRelayActions(w http.ResponseWriter, r *http.Request) {
	relayID := chi.URLParam(r, "id")
	userID := GetUserID(r)

	var req models.UpdateRelayActionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid request body", slog.String("error", err.Error()))
		h.respondError(w, http.StatusBadRequest, "Invalid JSON body", "INVALID_JSON")
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
		if !actions.IsValidType(action.ActionType) {
			h.respondError(w, http.StatusBadRequest,
				"Unknown action type '"+action.ActionType+"' at index "+strconv.Itoa(i),
				"VALIDATION_ERROR")
			return
		}
		if action.Config == nil {
			h.respondError(w, http.StatusBadRequest,
				"Config is required for action at index "+strconv.Itoa(i),
				"VALIDATION_ERROR")
			return
		}
		if err := actions.ValidateConfig(action.ActionType, action.Config); err != nil {
			h.respondError(w, http.StatusBadRequest,
				"Invalid config for action at index "+strconv.Itoa(i)+": "+err.Error(),
				"VALIDATION_ERROR")
			return
		}
	}

	relay, err := h.store.UpdateRelayActions(r.Context(), relayID, userID, req.Actions)
	if err != nil {
		if errors.Is(err, store.ErrRelayNotFound) {
			h.respondError(w, http.StatusNotFound, "Relay not found", "NOT_FOUND")
			return
		}
		h.logger.Error("failed to update relay actions",
			slog.String("relay_id", relayID),
			slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to update actions", "DB_ERROR")
		return
	}
	relay.Relay.WebhookURL = h.hooksURL + relay.Relay.WebhookPath

	h.logger.Info("relay actions updated",
		slog.String("relay_id", relayID),
		slog.Int("action_count", len(relay.Actions)))

	h.respondSuccess(w, http.StatusOK, "Relay actions updated successfully", relay)
}

func (h *Handler) DeleteRelay(w http.ResponseWriter, r *http.Request) {
	relayID := chi.URLParam(r, "id")
	userID := GetUserID(r)
	err := h.store.DeleteRelay(r.Context(), relayID, userID)
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

func (h *Handler) TriggerRelay(w http.ResponseWriter, r *http.Request) {
	relayID := chi.URLParam(r, "id")
	userID := GetUserID(r)

	relay, err := h.store.GetRelay(r.Context(), relayID, userID)
	if err != nil {
		if errors.Is(err, store.ErrRelayNotFound) {
			h.respondError(w, http.StatusNotFound, "Relay not found", "NOT_FOUND")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "Failed to fetch relay", "DB_ERROR")
		return
	}
	if relay.TriggerType != models.TriggerManual {
		h.respondError(w, http.StatusBadRequest,
			"This relay does not support manual triggering", "WRONG_TRIGGER_TYPE")
		return
	}
	if !relay.IsActive {
		h.respondError(w, http.StatusBadRequest, "Relay is not active", "RELAY_INACTIVE")
		return
	}

	var payload map[string]any
	_ = json.NewDecoder(r.Body).Decode(&payload)
	if payload == nil {
		payload = map[string]any{}
	}

	if err := h.publisher.PublishManualTrigger(r.Context(), relayID, payload); err != nil {
		h.logger.Error("failed to publish manual trigger",
			slog.String("relay_id", relayID),
			slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to trigger relay", "PUBLISH_ERROR")
		return
	}

	h.logger.Info("manual trigger fired", slog.String("relay_id", relayID), slog.String("user_id", userID))
	h.respondSuccess(w, http.StatusAccepted, "Relay triggered", map[string]string{"relay_id": relayID})
}

// HEALTH CHECK

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "hermes-core",
	})
}

// SECRETS API

func (h *Handler) CreateSecret(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid JSON body", "INVALID_JSON")
		return
	}

	req.UserID = GetUserID(r)

	if strings.TrimSpace(req.Name) == "" {
		h.respondError(w, http.StatusBadRequest, "name is required", "VALIDATION_ERROR")
		return
	}
	if strings.TrimSpace(req.Value) == "" {
		h.respondError(w, http.StatusBadRequest, "value is required", "VALIDATION_ERROR")
		return
	}
	secret, err := h.secretStore.Create(r.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			h.respondError(w, http.StatusConflict, err.Error(), "DUPLICATE_SECRET")
			return
		}
		h.logger.Error("failed to create secret", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to create secret", "DB_ERROR")
		return
	}
	h.logger.Info("secret created", slog.String("secret_id", secret.ID),
		slog.String("user_id", secret.UserID),
		slog.String("name", secret.Name))
	h.respondSuccess(w, http.StatusCreated, "Secret created", secret)
}

func (h *Handler) ListSecrets(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)

	secrets, err := h.secretStore.ListByUser(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to list secrets", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to list secrets", "DB_ERROR")
		return
	}

	h.respondSuccess(w, http.StatusOK, "", secrets)
}

func (h *Handler) DeleteSecret(w http.ResponseWriter, r *http.Request) {
	secretID := chi.URLParam(r, "id")
	userID := GetUserID(r)

	err := h.secretStore.Delete(r.Context(), userID, secretID)
	if err != nil {
		if errors.Is(err, store.ErrSecretNotFound) {
			h.respondError(w, http.StatusNotFound, "Secret not found", "NOT_FOUND")
			return
		}
		h.logger.Error("failed to delete secret", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to delete secret", "DB_ERROR")
		return
	}
	h.respondSuccess(w, http.StatusOK, "Secret deleted", map[string]string{"deleted_id": secretID})
}
