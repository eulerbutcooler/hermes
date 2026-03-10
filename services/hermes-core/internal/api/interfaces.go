package api

import (
	"context"
	"log/slog"
	"time"

	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/oauth"
	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/models"
)

type EventPublisher interface {
	PublishManualTrigger(ctx context.Context, relayID string, payload map[string]any) error
}

type RelayStorer interface {
	CreateRelay(ctx context.Context, req models.CreateRelayRequest) (*models.RelayWithActions, error)
	GetAllRelays(ctx context.Context, userID string) ([]models.Relay, error)
	GetRelay(ctx context.Context, relayID, userID string) (*models.RelayWithActions, error)
	UpdateRelay(ctx context.Context, relayID, userID string, req models.UpdateRelayRequest) (*models.Relay, error)
	UpdateRelayActions(ctx context.Context, relayID, userID string, actions []models.CreateRelayActionInput) (*models.RelayWithActions, error)
	DeleteRelay(ctx context.Context, relayID, userID string) error
	GetExecutions(ctx context.Context, relayID, userID string, limit int) ([]models.Execution, error)
	GetExecutionSteps(ctx context.Context, executionID, userID string) ([]models.ExecutionStep, error)
	DeleteExecution(ctx context.Context, executionID, userID string) error
}

type SecretStorer interface {
	Create(ctx context.Context, req models.CreateSecretRequest) (*models.Secret, error)
	ListByUser(ctx context.Context, userID string) ([]models.Secret, error)
	Delete(ctx context.Context, userID, secretID string) error
}

type UserStorer interface {
	CreateUser(ctx context.Context, username, email, passwordHash string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

type ConnectionStorer interface {
	Upsert(ctx context.Context, userID, provider, accountEmail, accessToken, refreshToken, scopes string, expiry time.Time) (*models.Connection, error)
	ListByUser(ctx context.Context, userID string) ([]models.Connection, error)
	Delete(ctx context.Context, userID, connectionID string) error
}

type Handler struct {
	store           RelayStorer
	secretStore     SecretStorer
	userStore       UserStorer
	connectionStore ConnectionStorer
	oauthProviders  map[string]oauth.Provider
	stateCodec      *oauth.StateCodec
	logger          *slog.Logger
	hooksURL        string
	jwtSecret       string
	publisher       EventPublisher
	frontendURL     string
}

func NewHandler(
	s RelayStorer,
	ss SecretStorer,
	us UserStorer,
	cs ConnectionStorer,
	providers map[string]oauth.Provider,
	stateCodec *oauth.StateCodec,
	jwtSecret string,
	logger *slog.Logger,
	publisher EventPublisher,
	hooksURL string,
	frontendURL string,
) *Handler {
	return &Handler{
		store:           s,
		secretStore:     ss,
		userStore:       us,
		connectionStore: cs,
		oauthProviders:  providers,
		stateCodec:      stateCodec,
		jwtSecret:       jwtSecret,
		logger:          logger,
		publisher:       publisher,
		hooksURL:        hooksURL,
		frontendURL:     frontendURL,
	}
}
