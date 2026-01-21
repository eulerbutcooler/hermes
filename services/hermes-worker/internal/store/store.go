package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RelayInstruction struct {
	RelayID    string
	ActionType string
	Config     map[string]any
}

type Store struct {
	db *pgxpool.Pool
}

var ErrRelayNotFound = errors.New("relay not found")

func NewStore(dbURL string) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to db: %w", err)
	}
	return &Store{db: pool}, nil
}

func (s *Store) GetRelayInstructions(ctx context.Context, relayID string) (*RelayInstruction, error) {
	query := `SELECT r.id, a.action_type, a.config
	FROM relays r
	JOIN relay_actions a ON r.id=a.relay_id
	WHERE r.id=$1 AND r.active=true`
	var ri RelayInstruction
	var configBytes []byte
	err := s.db.QueryRow(ctx, query, relayID).Scan(&ri.RelayID, &ri.ActionType, &configBytes)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRelayNotFound
		}
		return nil, fmt.Errorf("Db error: %w", err)
	}
	if err := json.Unmarshal(configBytes, &ri.Config); err != nil {
		return nil, fmt.Errorf("failed to parse config json: %w", err)
	}

	return &ri, nil
}

func (s *Store) LogExecution(ctx context.Context, relayID string, status string, details string) error {
	query := `INSERT INTO execution_logs(relay_id,status,details,executed_at)
	VALUES($1,$2,$3, NOW())`
	_, err := s.db.Exec(ctx, query, relayID, status, details)
	if err != nil {
		return fmt.Errorf("Failed to write execution log: %w", err)
	}
	return nil
}
