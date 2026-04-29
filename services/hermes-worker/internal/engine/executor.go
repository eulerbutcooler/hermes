package engine

import (
	"context"
	"encoding/json"
)

type StepOutput struct {
	ActionType string          `json:"action_type"`
	NodeID     string          `json:"node_id"`
	Output     json.RawMessage `json:"output"`
}

type ActionExecutor interface {
	Execute(ctx context.Context, config map[string]any, payload []byte, prevOutputs map[string]StepOutput) (json.RawMessage, error)
}
