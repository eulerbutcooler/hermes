package queue

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/eulerbutcooler/hermes-hooks/internal/api"

	"github.com/nats-io/nats.go"
)

type NatsQueue struct {
	js nats.JetStreamContext
}

var _ api.EventProducer = (*NatsQueue)(nil)

func NewNatsQueue(url string) (*NatsQueue, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("nats connect error: %w", err)
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("jetsream init error: %w", err)
	}
	streamName := "EVENTS"
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{"events.*"},
	})
	if err != nil {
		log.Printf("Stream %s might already exist: %v", streamName, err)
	}
	return &NatsQueue{js: js}, nil
}

func (q *NatsQueue) Publish(relayID string, event api.ExecutionEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	subject := fmt.Sprintf("events.%s", relayID)
	_, err = q.js.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("nats publish error: %w", err)
	}
	return nil
}
