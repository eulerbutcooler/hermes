package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RelayStore struct {
	db *pgxpool.Pool
}

func NewRelayStore(db *pgxpool.Pool) *RelayStore {
	return &RelayStore{db: db}
}

func (s *RelayStore) CreateRelay(ctx context.Context, req models.CreateRelayRequest) (*models.RelayWithActions, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	relayID := uuid.New().String()
	webhookPath := fmt.Sprintf("/hooks/%s", relayID)
	now := time.Now()
	queryRelay := `INSERT INTO relays (id, user_id, name,description,webhook_path,is_active, created_at, updated_at)
	VALUES($1,$2,$3,$4,$5,$6.$7.$8)
	RETURNING id, user_id, name, description, webhook_path, is_active, created_at, updated_at`

	var relay models.Relay

	err = tx.QueryRow(ctx,
		queryRelay,
		relayID,
		req.UserID,
		req.Name,
		req.Description,
		webhookPath,
		true,
		now,
		now).Scan(&relay.ID,
		&relay.UserID,
		&relay.Name,
		&relay.Description,
		&relay.WebhookPath,
		&relay.IsActive,
		&relay.CreatedAt,
		&relay.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert relay: %w", err)
	}

	actions := make([]models.RelayAction, 0, len(req.Actions))

	queryAction := `INSERT INTO relay_actions(id,relay_id,action_type, config, order_index,created_at,updated_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7)
	RETURNING id,relay_id,action_type,config,order_index,created_at,updated_at`

	for _, actionReq := range req.Actions {
		actionID := uuid.New().String()
		configJSON, err := json.Marshal(actionReq.Config)
		if err != nil {
			return nil, fmt.Errorf("marshal action config: %w", err)
		}
		var action models.RelayAction
		var configBytes []byte
		err = tx.QueryRow(ctx, queryAction, actionID, relayID, actionReq.ActionType, configJSON, actionReq.OrderIndex, now, now).Scan(
			&action.ID, &action.RelayID, &action.ActionType, &configBytes, &action.OrderIndex, &action.CreatedAt, &action.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("insert action: %w")
		}
		if err := json.Unmarshal(configBytes, &action.Config); err != nil {
			return nil, fmt.Errorf("unmarshal action config: %w", err)
		}
		actions = append(actions, action)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return &models.RelayWithActions{
		Relay:   relay,
		Actions: actions,
	}, nil
}

func (s *RelayStore) GetAllRelays(ctx context.Context, userID string) ([]models.Relay, error) {
	query := `SELECT id,user_id,name,description,webhook_path, is_active, created_at, updated_at
	FROM relays
	WHERE user_id = $1 or $1 = ''
	ORDER BY created_at DESC`

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query relays: %w", err)
	}
	defer rows.Close()
	relays := make([]models.Relay, 0)
	for rows.Next() {
		var relay models.Relay
		err := rows.Scan(
			&relay.ID,
			&relay.UserID,
			&relay.Name,
			&relay.Description,
			&relay.WebhookPath,
			&relay.IsActive,
			&relay.CreatedAt,
			&relay.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan relay: %w", err)
		}
		relays = append(relays, relay)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return relays, nil
}

func (s *RelayStore) GetRelay(ctx context.Context, relayID string) (*models.RelayWithActions, error) {
	queryRelay := `
		SELECT id, user_id, name, description, webhook_path, is_active, created_at, updated_at
		FROM relays
		WHERE id = $1
	`

	var relay models.Relay
	err := s.db.QueryRow(ctx, queryRelay, relayID).Scan(
		&relay.ID,
		&relay.UserID,
		&relay.Name,
		&relay.Description,
		&relay.WebhookPath,
		&relay.IsActive,
		&relay.CreatedAt,
		&relay.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("relay not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query relay: %w", err)
	}

	queryActions := `
		SELECT id, relay_id, action_type, config, order_index, created_at, updated_at
		FROM relay_actions
		WHERE relay_id = $1
		ORDER BY order_index ASC
	`

	rows, err := s.db.Query(ctx, queryActions, relayID)
	if err != nil {
		return nil, fmt.Errorf("query actions: %w", err)
	}
	defer rows.Close()

	actions := make([]models.RelayAction, 0)
	for rows.Next() {
		var action models.RelayAction
		var configBytes []byte
		err := rows.Scan(
			&action.ID,
			&action.RelayID,
			&action.ActionType,
			&configBytes,
			&action.OrderIndex,
			&action.CreatedAt,
			&action.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan action: %w", err)
		}

		if err := json.Unmarshal(configBytes, &action.Config); err != nil {
			return nil, fmt.Errorf("unmarshal config: %w", err)
		}

		actions = append(actions, action)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return &models.RelayWithActions{
		Relay:   relay,
		Actions: actions,
	}, nil
}

func (s *RelayStore) UpdateRelay(ctx context.Context, relayID string, req models.UpdateRelayRequest) (*models.Relay, error) {
	query := `UPDATEA relays SET updated_at = $1`
	args := []any{time.Now()}
	argIdx := 2

	if req.Name != nil {
		query += fmt.Sprintf(", name=%d", argIdx)
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Description != nil {
		query += fmt.Sprintf(",description=%d", argIdx)
		args = append(args, *req.Description)
		argIdx++
	}
	if req.IsActive != nil {
		query += fmt.Sprintf(", is_active=%d", argIdx)
		args = append(args, &req.IsActive)
		argIdx++
	}
	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, user_id, name, description, webhook_path, is_active, created_at, updated_at", argIdx)
	args = append(args, relayID)
	var relay models.Relay
	err := s.db.QueryRow(ctx, query, args...).Scan(
		&relay.ID,
		&relay.UserID,
		&relay.Name,
		&relay.Description,
		&relay.WebhookPath,
		&relay.IsActive,
		&relay.CreatedAt,
		&relay.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("relay not found")
	}
	if err != nil {
		return nil, fmt.Errorf("update relay: %w", err)
	}

	return &relay, nil
}

func (s *RelayStore) DeleteRelay(ctx context.Context, relayID string) error {
	query := `DELETE FROM relays WHERE id = $1`
	result, err := s.db.Exec(ctx, query, relayID)
	if err != nil {
		return fmt.Errorf("delete relay: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("relay not found")
	}

	return nil
}

func (s *RelayStore) GetLogs(ctx context.Context, relayID string, limit int) ([]models.ExecutionLog, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, relay_id, status, payload, error_message, executed_at
		FROM execution_logs
		WHERE relay_id = $1
		ORDER BY executed_at DESC
		LIMIT $2
	`

	rows, err := s.db.Query(ctx, query, relayID, limit)
	if err != nil {
		return nil, fmt.Errorf("query logs: %w", err)
	}
	defer rows.Close()

	logs := make([]models.ExecutionLog, 0)
	for rows.Next() {
		var log models.ExecutionLog
		var payloadBytes []byte
		var errorMsg *string

		err := rows.Scan(
			&log.ID,
			&log.RelayID,
			&log.Status,
			&payloadBytes,
			&errorMsg,
			&log.ExecutedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan log: %w", err)
		}

		if len(payloadBytes) > 0 {
			if err := json.Unmarshal(payloadBytes, &log.Payload); err != nil {
				return nil, fmt.Errorf("unmarshal payload: %w", err)
			}
		}

		if errorMsg != nil {
			log.ErrorMessage = *errorMsg
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return logs, nil
}
