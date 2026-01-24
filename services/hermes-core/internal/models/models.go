package models

import "time"

type CreateRelayRequest struct {
	Name        string                   `json:"name"`
	UserID      string                   `json:"user_id"`
	Description string                   `json:"description"`
	Actions     []CreateRelayActionInput `json:"actions"`
}

type CreateRelayActionInput struct {
	ActionType string         `json:"action_type"`
	Config     map[string]any `json:"config"`
	OrderIndex int            `json:"order_index"`
}

type UpdateRelayRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

type Relay struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	WebhookPath string    `json:"webhook_path"`
	WebhookURL  string    `json:"webhook_url"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RelayWithActions struct {
	Relay
	Actions []RelayAction `json:"actions"`
}

type RelayAction struct {
	ID         string         `json:"id"`
	RelayID    string         `json:"relay_id"`
	ActionType string         `json:"action_type"`
	Config     map[string]any `json:"config"`
	OrderIndex int            `json:"order_index"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type ExecutionLog struct {
	ID           string         `json:"id"`
	RelayID      string         `json:"relay_id"`
	Status       string         `json:"status"`
	Payload      map[string]any `json:"payload,omitempty"`
	ErrorMessage string         `json:"error_message,omitempty"`
	ExecutedAt   time.Time      `json:"executed_at"`
}

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}
