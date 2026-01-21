package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/logger"
	"github.com/go-chi/chi/v5"
)

// MockProducer satisfies the EventProducer interface
type MockProducer struct {
	LastRelayID string
}

func (m *MockProducer) Publish(zapID string, event ExecutionEvent) error {
	m.LastRelayID = zapID
	return nil
}

func TestHandleWebhook(t *testing.T) {
	mockQueue := &MockProducer{}
	testLogger := logger.New("hermes-hooks-test", "test", "debug")

	handler := NewHandler(mockQueue, testLogger)
	// Router to ensure URLParams are passed correctly
	r := chi.NewRouter()
	r.Post("/hooks/{relayID}", handler.HandleWebhook)

	// Request creation
	body := []byte(`{"test":"data"}`)
	req, _ := http.NewRequest("POST", "/hooks/test_relay_123", bytes.NewBuffer(body))

	// Record response
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Handler failed with status %d. Body: %s", rr.Code, rr.Body.String())
	}

	if mockQueue.LastRelayID != "test_relay_123" {
		t.Errorf("Expected RelayID 'test_zap_123', got '%s'", mockQueue.LastRelayID)
	}
}
