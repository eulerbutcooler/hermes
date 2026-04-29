package debug

import (
	"context"
	"encoding/json"
	"log"

	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/engine"
)

type LogExecutor struct{}

func New() *LogExecutor {
	return &LogExecutor{}
}

func (l *LogExecutor) Execute(ctx context.Context, config map[string]any, payload []byte, _ map[string]engine.StepOutput) (json.RawMessage, error) {
	prefix, _ := config["prefix"].(string)
	if prefix == "" {
		prefix = "DEBUG_LOG"
	}
	log.Printf("[%s] Payload Received: %s", prefix, string(payload))
	return json.RawMessage(payload), nil
}
