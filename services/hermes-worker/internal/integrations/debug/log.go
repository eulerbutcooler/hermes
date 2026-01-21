package debug

import (
	"context"
	"log"
)

type LogExecutor struct{}

func New() *LogExecutor {
	return &LogExecutor{}
}

func (l *LogExecutor) Execute(ctx context.Context, config map[string]any, payload []byte) error {
	prefix, _ := config["prefix"].(string)
	if prefix == "" {
		prefix = "DEBUG_LOG"
	}
	log.Printf("[%s] Payload Received: %s", prefix, string(payload))
	return nil
}
