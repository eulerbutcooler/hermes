package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RelayAction struct {
	OrderIndex int
	ActionType string
	Config     map[string]any
}

type Store struct {
	db *pgxpool.Pool
}

var (
	ErrRelayNotFound = errors.New("relay not found")
	ErrNoActions     = errors.New("no actions configured for relay")
)

func NewStore(dbURL string) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to db: %w", err)
	}
	return &Store{db: pool}, nil
}

func (s *Store) GetRelayActions(ctx context.Context, relayID string) ([]RelayAction, error) {
	query := `SELECT a.action_type, a.config, a.order_index
	FROM relays r
	JOIN relay_actions a ON r.id=a.relay_id
	WHERE r.id=$1 AND r.is_active=true
	ORDER BY a.order_index ASC`

	rows, err := s.db.Query(ctx, query, relayID)
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}
	defer rows.Close()

	actions := make([]RelayAction, 0)
	for rows.Next() {
		var act RelayAction
		var configBytes []byte
		if err := rows.Scan(&act.ActionType, &configBytes, &act.OrderIndex); err != nil {
			return nil, fmt.Errorf("scan action: %w", err)
		}
		if err := json.Unmarshal(configBytes, &act.Config); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
		actions = append(actions, act)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	if len(actions) == 0 {
		return nil, ErrNoActions
	}
	return actions, nil
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
