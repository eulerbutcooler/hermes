package models

type CreateRelayRequest struct {
	Name       string         `json:"name"`
	UserID     int            `json:"user_id"`
	ActionType string         `json:"action_type"`
	Config     map[string]any `json:"config"`
}

type Relay struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
}
