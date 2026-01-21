package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eulerbutcooler/hermes-core/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExecutionLog struct {
	ID         int       `json:"id"`
	Status     string    `json:"status"`
	Details    string    `json:"details"`
	ExecutedAt time.Time `json:"executed_at"`
}

type RelayStore struct {
	db *pgxpool.Pool
}

func NewRelayStore(db *pgxpool.Pool) *RelayStore {
	return &RelayStore{db: db}
}

func (s *RelayStore) CreateRelay(ctx context.Context, req models.CreateRelayRequest) (string, error) {
	relayID := uuid.New()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	queryRelay := `INSERT INTO relays (id, user_id, name) VALUES($1,$2,$3)`
	_, err = tx.Exec(ctx, queryRelay, relayID, req.UserID, req.Name)
	if err != nil {
		return "", fmt.Errorf("Failed to insert relay: %w", err)
	}
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return "", err
	}

	queryAction := `INSERT INTO relay_actions (relay_id, action_type, config) VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, queryAction, relayID, req.ActionType, configJSON)
	if err != nil {
		return "", fmt.Errorf("failed to insert action: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return relayID.String(), nil
}

func (s *RelayStore) GetLogs(ctx context.Context, relayID string) ([]ExecutionLog, error) {
	query := `SELECT id, status, details, executed_at
	FROM execution_logs
	WHERE relay_id=$1
	ORDER BY executed_at DESC
	LIMIT 50`
	rows, err := s.db.Query(ctx, query, relayID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()
	var logs []ExecutionLog
	for rows.Next() {
		var l ExecutionLog
		if err := rows.Scan(&l.ID, &l.Status, &l.Details, &l.ExecutedAt); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	if logs == nil {
		logs = []ExecutionLog{}
	}
	return logs, nil
}
