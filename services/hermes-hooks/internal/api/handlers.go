package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ExecutionEvent struct {
	EventID    string          `json:"event_id"`
	RelayID    string          `json:"relay_id"`
	Payload    json.RawMessage `json:"payload"`
	ReceivedAt time.Time       `json:"received_at"`
}

type EventProducer interface {
	Publish(relayID string, event ExecutionEvent) error
}

type Handler struct {
	producer EventProducer
	logger   *slog.Logger
}

func NewHandler(p EventProducer, logger *slog.Logger) *Handler {
	return &Handler{producer: p, logger: logger}
}

func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	relayID := chi.URLParam(r, "relayID")
	if relayID == "" {
		h.logger.Warn("webhook request missing relay ID",
			slog.String("path", r.URL.Path),
		)
		http.Error(w, "Relay ID is required", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		h.logger.Error("failed to read request body",
			slog.String("relay_id", relayID),
			slog.String("error", err.Error()),
		)
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	eventID := r.Header.Get("X-Event-ID")
	if eventID == "" {
		eventID = r.URL.Query().Get("event_id")
	}
	if eventID == "" {
		eventID = uuid.New().String()
	}

	h.logger.Debug("webhook received",
		slog.String("relay_id", relayID),
		slog.Int("payload_size", len(body)),
		slog.String("content_type", r.Header.Get("Content-Type")),
	)

	event := ExecutionEvent{
		EventID:    eventID,
		RelayID:    relayID,
		Payload:    body,
		ReceivedAt: time.Now(),
	}
	if err := h.producer.Publish(relayID, event); err != nil {
		h.logger.Error("failed to publish event",
			slog.String("relay_id", relayID),
			slog.String("error", err.Error()),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("webhook queued successfully",
		slog.String("relay_id", relayID),
		slog.String("event_id", eventID),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf(`{"status":"queued", "event_id":"%s"}`, eventID)))
}
