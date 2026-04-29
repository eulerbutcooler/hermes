package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/encryptor"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CronRelay struct {
	ID            string
	TriggerConfig map[string]any
}

type RelayAction struct {
	NodeID     string
	ActionType string
	Config     map[string]any
}

type RelayEdge struct {
	ParentNodeID string
	ChildNodeID  string
	Condition    map[string]any
}

type Store struct {
	db        *pgxpool.Pool
	encryptor *encryptor.Encryptor
}

type Execution struct {
	ID string
}

type ExecutionStep struct {
	ID string
}

var (
	ErrRelayNotFound      = errors.New("relay not found")
	ErrNoActions          = errors.New("no actions configured for relay")
	ErrSecretNotFound     = errors.New("secret not found")
	ErrConnectionNotFound = errors.New("connection not found")
)

func NewStore(dbURL string, enc *encryptor.Encryptor) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to db: %w", err)
	}
	return &Store{db: pool, encryptor: enc}, nil
}

func (s *Store) GetRelayGraph(ctx context.Context, relayID string) ([]RelayAction, []RelayEdge, error) {
	actionsQuery := `SELECT a.node_id, a.action_type, a.config
	FROM relays r
	JOIN relay_actions a ON r.id=a.relay_id
	WHERE r.id = $1 AND r.is_active=true
	`
	rows, err := s.db.Query(ctx, actionsQuery, relayID)
	if err != nil {
		return nil, nil, fmt.Errorf("db query actions error: %w", err)
	}
	defer rows.Close()

	actions := make([]RelayAction, 0)
	for rows.Next() {
		var act RelayAction
		var configBytes []byte
		if err := rows.Scan(&act.NodeID, &act.ActionType, &configBytes); err != nil {
			return nil, nil, fmt.Errorf("scan action: %w", err)
		}
		if err := json.Unmarshal(configBytes, &act.Config); err != nil {
			return nil, nil, fmt.Errorf("parse config: %w", err)
		}
		actions = append(actions, act)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("actions rows error: %w", err)
	}
	if len(actions) == 0 {
		return nil, nil, ErrNoActions
	}

	edgesQuery := `
	SELECT parent_node_id, child_node_id, condition
	FROM relay_edges
	WHERE relay_id = $1`

	edgeRows, err := s.db.Query(ctx, edgesQuery, relayID)
	if err != nil {
		return nil, nil, fmt.Errorf("db query edges error: %w", err)
	}

	defer edgeRows.Close()

	edges := make([]RelayEdge, 0)
	for edgeRows.Next() {
		var edge RelayEdge
		var conditionBytes []byte
		if err := edgeRows.Scan(&edge.ParentNodeID, &edge.ChildNodeID, &conditionBytes); err != nil {
			return nil, nil, fmt.Errorf("scan edge: %w", err)
		}
		if len(conditionBytes) > 0 {
			if err := json.Unmarshal(conditionBytes, &edge.Condition); err != nil {
				return nil, nil, fmt.Errorf("parse edge condition: %w", err)
			}
		}
		edges = append(edges, edge)
	}
	if err := edgeRows.Err(); err != nil {
		return nil, nil, fmt.Errorf("edges rows error: %w", err)
	}
	return actions, edges, nil
}

func (s *Store) GetRelayOwner(ctx context.Context, relayID string) (string, error) {
	var userID string
	err := s.db.QueryRow(ctx, `SELECT user_id FROM relays WHERE id = $1`, relayID).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrRelayNotFound
	}
	if err != nil {
		return "", fmt.Errorf("query relay owner: %w", err)
	}
	return userID, nil
}

func (s *Store) RegisterEvent(ctx context.Context, relayID, eventID string) (bool, error) {
	if eventID == "" {
		return true, nil
	}
	query := `INSERT INTO processed_events (relay_id, event_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`
	tag, err := s.db.Exec(ctx, query, relayID, eventID)
	if err != nil {
		return false, fmt.Errorf("dedupe insert failed: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (s *Store) LogExecution(ctx context.Context, relayID string, eventID string, status string, details string, payload []byte) error {
	query := `INSERT INTO execution_logs(relay_id, event_id, status,payload,error_message,executed_at)
	VALUES($1,$2,$3,$4,$5,NOW())`

	var payloadJSON any
	if len(payload) > 0 {
		payloadJSON = json.RawMessage(payload)
	}

	var errorMessage any
	if status != "success" && details != "" {
		errorMessage = details
	}

	_, err := s.db.Exec(ctx, query, relayID, eventID, status, payloadJSON, errorMessage)
	if err != nil {
		return fmt.Errorf("failed to write execution log: %w", err)
	}
	return nil
}

func (s *Store) ResolveSecret(ctx context.Context, userID, secretName string) (string, error) {
	var encrypted string
	err := s.db.QueryRow(ctx,
		`SELECT value FROM secrets WHERE user_id = $1 AND name = $2`,
		userID, secretName,
	).Scan(&encrypted)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("%w: %s", ErrSecretNotFound, secretName)
	}
	if err != nil {
		return "", fmt.Errorf("resolve secret: %w", err)
	}
	plaintext, err := s.encryptor.Decrypt(encrypted)
	if err != nil {
		return "", fmt.Errorf("decrypt secret %q: %w", secretName, err)
	}
	return plaintext, nil
}

func (s *Store) GetConnection(ctx context.Context, connectionID string) (provider, accessToken, refreshToken, accountEmail string, expiry time.Time, err error) {
	var encAccess, encRefresh string
	err = s.db.QueryRow(ctx, `SELECT provider, access_token, refresh_token, account_email, token_expiry FROM connections WHERE id = $1`,
		connectionID).Scan(&provider, &encAccess, &encRefresh, &accountEmail, &expiry)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", "", "", time.Time{}, fmt.Errorf("%w: %s", ErrConnectionNotFound, connectionID)
	}
	if err != nil {
		return "", "", "", "", time.Time{}, fmt.Errorf("query connection: %w", err)
	}

	accessToken, err = s.encryptor.Decrypt(encAccess)
	if err != nil {
		return "", "", "", "", time.Time{}, fmt.Errorf("decrypt access token: %w", err)
	}
	refreshToken, err = s.encryptor.Decrypt(encRefresh)
	if err != nil {
		return "", "", "", "", time.Time{}, fmt.Errorf("decrypt refresh token: %w", err)
	}

	return provider, accessToken, refreshToken, accountEmail, expiry, nil
}

func (s *Store) UpdateConnectionTokens(ctx context.Context, connectionID, accessToken, refreshToken string, expiry time.Time) error {
	encAccess, err := s.encryptor.Encrypt(accessToken)
	if err != nil {
		return fmt.Errorf("encrypt access token: %w", err)
	}
	encRefresh, err := s.encryptor.Encrypt(refreshToken)
	if err != nil {
		return fmt.Errorf("encrypt refresh token: %w", err)
	}

	_, err = s.db.Exec(ctx,
		`UPDATE connections SET access_token = $1, refresh_token = $2, token_expiry = $3, updated_at = NOW() WHERE id = $4`,
		encAccess, encRefresh, expiry, connectionID)
	return err
}

func (s *Store) CreateExecution(ctx context.Context, relayID, eventID string, triggerPayload []byte) (string, error) {
	query := `
		INSERT INTO executions (relay_id, event_id, status, trigger_payload, started_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id
	`

	var payloadJSON any
	if len(triggerPayload) > 0 {
		payloadJSON = json.RawMessage(triggerPayload)
	}

	var executionID string
	err := s.db.QueryRow(ctx, query, relayID, eventID, "running", payloadJSON).Scan(&executionID)
	if err != nil {
		return "", fmt.Errorf("create execution: %w", err)
	}
	return executionID, nil
}

func (s *Store) CompleteExecution(ctx context.Context, executionID, status, errorMessage string) error {
	query := `
		UPDATE executions
		SET status = $2,
		    error_message = $3,
		    finished_at = NOW()
		WHERE id = $1
	`
	_, err := s.db.Exec(ctx, query, executionID, status, nullableString(errorMessage))
	if err != nil {
		return fmt.Errorf("complete execution: %w", err)
	}
	return nil
}

func (s *Store) CreateExecutionStep(ctx context.Context, executionID string, nodeID string, actionType string, input map[string]any) (string, error) {
	query := `
		INSERT INTO execution_steps (execution_id, node_id, action_type, status, input, started_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id
	`

	var inputJSON any
	if input != nil {
		b, err := json.Marshal(input)
		if err != nil {
			return "", fmt.Errorf("marshal step input: %w", err)
		}
		inputJSON = json.RawMessage(b)
	}

	var stepID string
	err := s.db.QueryRow(ctx, query, executionID, nodeID, actionType, "running", inputJSON).Scan(&stepID)
	if err != nil {
		return "", fmt.Errorf("create execution step: %w", err)
	}
	return stepID, nil
}

func (s *Store) CompleteExecutionStep(ctx context.Context, stepID, status, errorMessage string, output []byte) error {
	query := `
		UPDATE execution_steps
		SET status = $2,
		    output = $3,
		    error_message = $4,
		    finished_at = NOW()
		WHERE id = $1
	`

	var outputJSON any
	if len(output) > 0 {
		outputJSON = json.RawMessage(output)
	}

	_, err := s.db.Exec(ctx, query, stepID, status, outputJSON, nullableString(errorMessage))
	if err != nil {
		return fmt.Errorf("complete execution step: %w", err)
	}
	return nil
}

func nullableString(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func (s *Store) GetCronRelaysDue(ctx context.Context) ([]CronRelay, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, trigger_config
		FROM relays
		WHERE trigger_type = 'cron'
		  AND is_active = true
		  AND next_run_at IS NOT NULL
		  AND next_run_at <= NOW()
	`)
	if err != nil {
		return nil, fmt.Errorf("query cron relays: %w", err)
	}
	defer rows.Close()

	var relays []CronRelay
	for rows.Next() {
		var r CronRelay
		var configBytes []byte
		if err := rows.Scan(&r.ID, &configBytes); err != nil {
			return nil, fmt.Errorf("scan cron relay: %w", err)
		}
		if err := json.Unmarshal(configBytes, &r.TriggerConfig); err != nil {
			return nil, fmt.Errorf("parse trigger_config: %w", err)
		}
		relays = append(relays, r)
	}
	return relays, rows.Err()
}

func (s *Store) UpdateRelayNextRun(ctx context.Context, relayID string, nextRunAt *time.Time) error {
	_, err := s.db.Exec(ctx, `
		UPDATE relays
		SET last_run_at = NOW(), next_run_at = $2, updated_at = NOW()
		WHERE id = $1
	`, relayID, nextRunAt)
	return err
}
