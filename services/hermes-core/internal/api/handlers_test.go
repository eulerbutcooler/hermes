package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/models"
	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

// Mock Relay Store records calls and returns configurable responses
type mockRelayStore struct {
	createRelayCalled bool
	createRelayReq    models.CreateRelayRequest
	createRelayResult *models.RelayWithActions
	createRelayErr    error

	getAllResult []models.Relay
	getAllErr    error

	getRelayResult *models.RelayWithActions
	getRelayErr    error

	updateRelayResult *models.Relay
	updateRelayErr    error

	updateRelayActionsResult *models.RelayWithActions
	updateRelayActionsErr    error

	deleteRelayErr     error
	deleteExecutionErr error

	getExecutionsResult []models.Execution
	getExecutionsErr    error

	getExecutionStepsResult []models.ExecutionStep
	getExecutionStepsErr    error
}

// RELAY TESTS

func (m *mockRelayStore) CreateRelay(_ context.Context, req models.CreateRelayRequest) (*models.RelayWithActions, error) {
	m.createRelayCalled = true
	m.createRelayReq = req
	return m.createRelayResult, m.createRelayErr
}

func (m *mockRelayStore) GetAllRelays(_ context.Context, _ string) ([]models.Relay, error) {
	return m.getAllResult, m.getAllErr
}

func (m *mockRelayStore) GetRelay(_ context.Context, _, _ string) (*models.RelayWithActions, error) {
	return m.getRelayResult, m.getRelayErr
}

func (m *mockRelayStore) UpdateRelay(_ context.Context, _, _ string, _ models.UpdateRelayRequest) (*models.Relay, error) {
	return m.updateRelayResult, m.updateRelayErr
}

func (m *mockRelayStore) UpdateRelayActions(_ context.Context, _, _ string, _ models.UpdateRelayActionsRequest) (*models.RelayWithActions, error) {
	return m.updateRelayActionsResult, m.updateRelayActionsErr
}

func (m *mockRelayStore) DeleteRelay(_ context.Context, _, _ string) error {
	return m.deleteRelayErr
}

func (m *mockRelayStore) GetExecutions(_ context.Context, _, _ string, _ int) ([]models.Execution, error) {
	return m.getExecutionsResult, m.getExecutionsErr
}

func (m *mockRelayStore) GetExecutionSteps(_ context.Context, _, _ string) ([]models.ExecutionStep, error) {
	return m.getExecutionStepsResult, m.getExecutionStepsErr
}

func (m *mockRelayStore) DeleteExecution(_ context.Context, _, _ string) error {
	return m.deleteExecutionErr
}

type mockSecretStore struct {
	createResult *models.Secret
	createErr    error
	listResult   []models.Secret
	listErr      error
	deleteErr    error
}

func (m *mockSecretStore) Create(_ context.Context, _ models.CreateSecretRequest) (*models.Secret, error) {
	return m.createResult, m.createErr
}

func (m *mockSecretStore) ListByUser(_ context.Context, _ string) ([]models.Secret, error) {
	return m.listResult, m.listErr
}

func (m *mockSecretStore) Delete(_ context.Context, _, _ string) error {
	return m.deleteErr
}

type mockUserStore struct {
	createUserResult *models.User
	createUserErr    error
	getUserResult    *models.User
	getUserErr       error
}

func (m *mockUserStore) CreateUser(_ context.Context, _, _, _ string) (*models.User, error) {
	return m.createUserResult, m.createUserErr
}

func (m *mockUserStore) GetUserByEmail(_ context.Context, _ string) (*models.User, error) {
	return m.getUserResult, m.getUserErr
}

type mockConnectionStore struct {
	upsertResult *models.Connection
	upsertErr    error
	listResult   []models.Connection
	listErr      error
	deleteErr    error
}

func (m *mockConnectionStore) Upsert(_ context.Context, _, _, _, _, _, _ string, _ time.Time) (*models.Connection, error) {
	return m.upsertResult, m.upsertErr
}

func (m *mockConnectionStore) ListByUser(_ context.Context, _ string) ([]models.Connection, error) {
	return m.listResult, m.listErr
}

func (m *mockConnectionStore) Delete(_ context.Context, _, _ string) error {
	return m.deleteErr
}

const testJWTsecret = "test-secret-key-for-jwt"

// newTestHandler creates a Handler wired to the give mock stores
func newTestHandler(rs *mockRelayStore, ss *mockSecretStore, us *mockUserStore) *Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewHandler(rs, ss, us, &mockConnectionStore{}, nil, nil, testJWTsecret, logger, nil, "", "")
}

// Creates a valid JWT for the given userID, simulating
// an authenticated request. This is to bypass the middleware
func generateJWT(userID string) string {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testJWTsecret))
	return signed
}

// Creates an HTTP request with a valid JWT in the Authorization header
// and injects the userID into the context
func authedRequest(method, path string, body []byte, userID string) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), userIDKey, userID)
	return req.WithContext(ctx)
}

// Helper function to unmarshal JSON response body
func decodeResponse(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to deconde response body:  %v\nbody: %s", err, rr.Body.String())
	}
	return resp
}

// AUTH HANDLER TESTS

// Verifies the happy path for user registeration
func TestRegister_Success(t *testing.T) {
	us := &mockUserStore{
		createUserResult: &models.User{
			ID:        "user-1",
			Username:  "aman",
			Email:     "aman@yadav.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	h := newTestHandler(&mockRelayStore{}, &mockSecretStore{}, us)

	// Valid registeration request body
	body, _ := json.Marshal(models.RegisterRequest{
		Username: "aman",
		Email:    "aman@yadav.com",
		Password: "yopierreyouwannacomeouthere",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Calling the handler directly
	h.Register(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	resp := decodeResponse(t, rr)
	if resp["success"] != true {
		t.Errorf("expected success=true, got %v", resp["success"])
	}
}

// Tests short passwords and missing fields
func TestRegister_ValidationErrors(t *testing.T) {
	h := newTestHandler(&mockRelayStore{}, &mockSecretStore{}, &mockUserStore{})
	cases := []struct {
		name string
		body models.RegisterRequest
	}{
		{"empty_username", models.RegisterRequest{Username: "", Email: "yo@po.com", Password: "12345678"}},
		{"empty_email", models.RegisterRequest{Username: "geetu", Email: "", Password: "12345678"}},
		{"short_password", models.RegisterRequest{Username: "geetu", Email: "g@t.com", Password: "short"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			h.Register(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d: %s", rr.Code, rr.Body.String())
			}
		})
	}
}

// Verifies the 409 conflict when email is taken
func TestRegister_DuplicateEmail(t *testing.T) {
	us := &mockUserStore{
		createUserErr: store.ErrEmailTaken,
	}
	h := newTestHandler(&mockRelayStore{}, &mockSecretStore{}, us)
	body, _ := json.Marshal(models.RegisterRequest{
		Username: "geeti",
		Email:    "git@ikea.com",
		Password: "mastikarnihai",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rr.Code, rr.Body.String())
	}
}

// Ensures that a malformed body returns 400
func TestLogin_InvalidJSO(t *testing.T) {
	h := newTestHandler(&mockRelayStore{}, &mockSecretStore{}, &mockUserStore{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// Verifies that a non-existent email returns 401
func TestLogin_UserNotFound(t *testing.T) {
	us := &mockUserStore{
		getUserErr: store.ErrUserNotFound,
	}
	h := newTestHandler(&mockRelayStore{}, &mockSecretStore{}, us)

	body, _ := json.Marshal(models.LoginRequest{
		Email:    "huh@what.com",
		Password: "koiipasshaikya",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}
}

// RELAY HANDLER TESTS

// Tests happy patyh for relay creation
func TestCreateRelay_Success(t *testing.T) {
	now := time.Now()
	rs := &mockRelayStore{
		createRelayResult: &models.RelayWithActions{
			Relay: models.Relay{
				ID:          "relay-1",
				UserID:      "user-1",
				Name:        "My Relay",
				Description: "Test relay",
				WebhookPath: "/hooks/relay-1",
				IsActive:    true,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			Actions: []models.RelayAction{
				{
					ID:         "action-1",
					RelayID:    "relay-1",
					ActionType: "debug_log",
					Config:     map[string]any{"prefix": "TEST"},
					NodeID:     "node_0",
					CreatedAt:  now,
					UpdatedAt:  now,
				},
			},
		},
	}
	h := newTestHandler(rs, &mockSecretStore{}, &mockUserStore{})

	// Valid create request with one debug_log action.
	body, _ := json.Marshal(map[string]any{
		"name":        "My Relay",
		"description": "Test relay",
		"actions": []map[string]any{
			{"action_type": "debug_log", "config": map[string]any{"prefix": "TEST"}, "node_id": "node_0"},
		},
	})

	req := authedRequest(http.MethodPost, "/api/v1/relays", body, "user-1")
	rr := httptest.NewRecorder()

	h.CreateRelay(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify the store was called.
	if !rs.createRelayCalled {
		t.Error("expected CreateRelay to be called on store")
	}

	// Verify the user_id was injected from context, not from the request body.
	if rs.createRelayReq.UserID != "user-1" {
		t.Errorf("expected user_id='user-1', got %q", rs.createRelayReq.UserID)
	}
}

// Tests that a relay without a name is rejected.
func TestCreateRelay_MissingName(t *testing.T) {
	h := newTestHandler(&mockRelayStore{}, &mockSecretStore{}, &mockUserStore{})

	body, _ := json.Marshal(map[string]any{
		"name": "",
		"actions": []map[string]any{
			{"action_type": "debug_log", "config": map[string]any{}, "node_id": "node_0"},
		},
	})

	req := authedRequest(http.MethodPost, "/api/v1/relays", body, "user-1")
	rr := httptest.NewRecorder()
	h.CreateRelay(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

// Tests that at least one action is required.
func TestCreateRelay_NoActions(t *testing.T) {
	h := newTestHandler(&mockRelayStore{}, &mockSecretStore{}, &mockUserStore{})

	body, _ := json.Marshal(map[string]any{
		"name":    "My Relay",
		"actions": []map[string]any{},
	})

	req := authedRequest(http.MethodPost, "/api/v1/relays", body, "user-1")
	rr := httptest.NewRecorder()
	h.CreateRelay(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

// Checks that an unregistered action type is rejected.
func TestCreateRelay_UnknownActionType(t *testing.T) {
	h := newTestHandler(&mockRelayStore{}, &mockSecretStore{}, &mockUserStore{})

	body, _ := json.Marshal(map[string]any{
		"name": "My Relay",
		"actions": []map[string]any{
			{"action_type": "nonexistent_integration", "config": map[string]any{}, "node_id": "node_0"},
		},
	})

	req := authedRequest(http.MethodPost, "/api/v1/relays", body, "user-1")
	rr := httptest.NewRecorder()
	h.CreateRelay(rr, req)

	// Should be 400 because the action type is unknown.
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

// Checks that the list endpoint returns relays with webhook URLs.
func TestGetAllRelays_Success(t *testing.T) {
	rs := &mockRelayStore{
		getAllResult: []models.Relay{
			{ID: "r1", UserID: "user-1", Name: "Relay 1", WebhookPath: "/hooks/r1", IsActive: true},
			{ID: "r2", UserID: "user-1", Name: "Relay 2", WebhookPath: "/hooks/r2", IsActive: false},
		},
	}
	h := newTestHandler(rs, &mockSecretStore{}, &mockUserStore{})

	req := authedRequest(http.MethodGet, "/api/v1/relays", nil, "user-1")
	rr := httptest.NewRecorder()
	h.GetAllRelays(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	resp := decodeResponse(t, rr)
	if resp["success"] != true {
		t.Errorf("expected success=true")
	}
}

// Checks that a missing relay returns 404.
func TestGetRelay_NotFound(t *testing.T) {
	rs := &mockRelayStore{
		getRelayErr: store.ErrRelayNotFound,
	}
	h := newTestHandler(rs, &mockSecretStore{}, &mockUserStore{})

	// We need chi's URL params, so set up a mini router.
	r := chi.NewRouter()
	r.Get("/api/v1/relays/{id}", h.GetRelay)

	req := authedRequest(http.MethodGet, "/api/v1/relays/nonexistent", nil, "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

// Checks that an empty update body is rejected.
func TestUpdateRelay_NoFields(t *testing.T) {
	h := newTestHandler(&mockRelayStore{}, &mockSecretStore{}, &mockUserStore{})

	// chi router for URL params
	r := chi.NewRouter()
	r.Put("/api/v1/relays/{id}", h.UpdateRelay)

	body, _ := json.Marshal(map[string]any{})
	req := authedRequest(http.MethodPut, "/api/v1/relays/relay-1", body, "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

// Tests the happy path for relay deletion.
func TestDeleteRelay_Success(t *testing.T) {
	rs := &mockRelayStore{
		deleteRelayErr: nil, // success
	}
	h := newTestHandler(rs, &mockSecretStore{}, &mockUserStore{})

	r := chi.NewRouter()
	r.Delete("/api/v1/relays/{id}", h.DeleteRelay)

	req := authedRequest(http.MethodDelete, "/api/v1/relays/relay-1", nil, "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

// Tests deletion of a non-existent relay.
func TestDeleteRelay_NotFound(t *testing.T) {
	rs := &mockRelayStore{
		deleteRelayErr: store.ErrRelayNotFound,
	}
	h := newTestHandler(rs, &mockSecretStore{}, &mockUserStore{})

	r := chi.NewRouter()
	r.Delete("/api/v1/relays/{id}", h.DeleteRelay)

	req := authedRequest(http.MethodDelete, "/api/v1/relays/nonexistent", nil, "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestGetExecutions_Success(t *testing.T) {
	now := time.Now()
	rs := &mockRelayStore{
		getExecutionsResult: []models.Execution{
			{
				ID:           "exec-1",
				RelayID:      "relay-1",
				EventID:      "event-1",
				Status:       "success",
				StartedAt:    now,
				FinishedAt:   &now,
				ErrorMessage: "",
			},
			{
				ID:           "exec-2",
				RelayID:      "relay-1",
				Status:       "failed",
				StartedAt:    now,
				FinishedAt:   &now,
				ErrorMessage: "step failed",
			},
		},
	}
	h := newTestHandler(rs, &mockSecretStore{}, &mockUserStore{})

	r := chi.NewRouter()
	r.Get("/api/v1/relays/{id}/executions", h.GetExecutions)

	req := authedRequest(http.MethodGet, "/api/v1/relays/relay-1/executions?limit=50", nil, "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	resp := decodeResponse(t, rr)
	if resp["success"] != true {
		t.Errorf("expected success=true, got %v", resp["success"])
	}

	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatalf("expected data to be an array, got %T", resp["data"])
	}
	if len(data) != 2 {
		t.Fatalf("expected 2 executions, got %d", len(data))
	}
}

func TestGetExecutionSteps_Success(t *testing.T) {
	now := time.Now()
	rs := &mockRelayStore{
		getExecutionStepsResult: []models.ExecutionStep{
			{
				ID:          "step-1",
				ExecutionID: "exec-1",
				NodeID:      "node_0",
				ActionType:  "http_request",
				Status:      "success",
				Input: map[string]any{
					"url":    "https://httpbin.org/get",
					"method": "GET",
				},
				Output: map[string]any{
					"status_code":  200,
					"content_type": "application/json",
				},
				StartedAt:  now,
				FinishedAt: &now,
			},
		},
	}
	h := newTestHandler(rs, &mockSecretStore{}, &mockUserStore{})

	r := chi.NewRouter()
	r.Get("/api/v1/executions/{executionId}/steps", h.GetExecutionSteps)

	req := authedRequest(http.MethodGet, "/api/v1/executions/exec-1/steps", nil, "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	resp := decodeResponse(t, rr)
	if resp["success"] != true {
		t.Errorf("expected success=true, got %v", resp["success"])
	}

	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatalf("expected data to be an array, got %T", resp["data"])
	}
	if len(data) != 1 {
		t.Fatalf("expected 1 execution step, got %d", len(data))
	}
}

// SECRET HANDLER TESTS

// Tests the happy path for secret creation.
func TestCreateSecret_Success(t *testing.T) {
	ss := &mockSecretStore{
		createResult: &models.Secret{
			ID:        "secret-1",
			UserID:    "user-1",
			Name:      "my_api_key",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	h := newTestHandler(&mockRelayStore{}, ss, &mockUserStore{})

	body, _ := json.Marshal(map[string]any{
		"name":  "my_api_key",
		"value": "supersecretvalue123",
	})

	req := authedRequest(http.MethodPost, "/api/v1/secrets", body, "user-1")
	rr := httptest.NewRecorder()
	h.CreateSecret(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

// Tests that a secret without a name is rejected.
func TestCreateSecret_MissingName(t *testing.T) {
	h := newTestHandler(&mockRelayStore{}, &mockSecretStore{}, &mockUserStore{})

	body, _ := json.Marshal(map[string]any{
		"name":  "",
		"value": "some-value",
	})

	req := authedRequest(http.MethodPost, "/api/v1/secrets", body, "user-1")
	rr := httptest.NewRecorder()
	h.CreateSecret(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

// Tests that a secret without a value is rejected.
func TestCreateSecret_MissingValue(t *testing.T) {
	h := newTestHandler(&mockRelayStore{}, &mockSecretStore{}, &mockUserStore{})

	body, _ := json.Marshal(map[string]any{
		"name":  "my_key",
		"value": "",
	})

	req := authedRequest(http.MethodPost, "/api/v1/secrets", body, "user-1")
	rr := httptest.NewRecorder()
	h.CreateSecret(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

// Verifies listing secrets for a user.
func TestListSecrets_Success(t *testing.T) {
	ss := &mockSecretStore{
		listResult: []models.Secret{
			{ID: "s1", UserID: "user-1", Name: "key1"},
			{ID: "s2", UserID: "user-1", Name: "key2"},
		},
	}
	h := newTestHandler(&mockRelayStore{}, ss, &mockUserStore{})

	req := authedRequest(http.MethodGet, "/api/v1/secrets", nil, "user-1")
	rr := httptest.NewRecorder()
	h.ListSecrets(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

// Verifies 404 when deleting a non-existent secret.
func TestDeleteSecret_NotFound(t *testing.T) {
	ss := &mockSecretStore{
		deleteErr: store.ErrSecretNotFound,
	}
	h := newTestHandler(&mockRelayStore{}, ss, &mockUserStore{})

	r := chi.NewRouter()
	r.Delete("/api/v1/secrets/{id}", h.DeleteSecret)

	req := authedRequest(http.MethodDelete, "/api/v1/secrets/nonexistent", nil, "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

// HEALTH CHECK TEST

// Verifies the health endpoint returns 200 with status "healthy".
func TestHealthCheck(t *testing.T) {
	h := newTestHandler(&mockRelayStore{}, &mockSecretStore{}, &mockUserStore{})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	h.HealthCheck(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	resp := decodeResponse(t, rr)
	if resp["status"] != "healthy" {
		t.Errorf("expected status=healthy, got %v", resp["status"])
	}
}
