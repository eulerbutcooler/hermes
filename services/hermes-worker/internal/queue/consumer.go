package queue

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/engine"
	"github.com/nats-io/nats.go"
)

type Consumer struct {
	js       nats.JetStream
	sub      *nats.Subscription
	jobQueue chan engine.Job
	logger   *slog.Logger
}

// Constructor pattern
// Initializes the NATS connection but doesnt start consuming right off
func NewConsumer(url string, jobQueue chan engine.Job, logger *slog.Logger) (*Consumer, error) {
	nc, err := nats.Connect(
		url,
		nats.MaxReconnects(10),
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			logger.Warn("NATS disconnected", slog.String("error", err.Error()))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Info("NATS reconnected")
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("nats connect error: %w", err)
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("jetsream init error: %w", err)
	}
	logger.Info("connected to NATS JetStream")
	return &Consumer{
		js:       js,
		jobQueue: jobQueue,
		logger:   logger,
	}, nil
}

// Consumes the messages by subscribing to NATS and processing messages async
func (c *Consumer) Start() error {
	c.logger.Info("starting NATS consumer",
		slog.String("subject", "events.>"),
		slog.String("consumer", "WORKER_CONSUMER"))
	sub, err := c.js.Subscribe("events.>",
		c.handleMessage,
		nats.Durable("WORKER_CONSUMER"),
		nats.ManualAck(),
		nats.AckWait(30*time.Second))
	if err != nil {
		return fmt.Errorf("subscription failed: %w", err)
	}
	c.sub = sub
	c.logger.Info("Worker consumer started, listening for events...")
	return nil
}

func (c *Consumer) handleMessage(msg *nats.Msg) {
	type Event struct {
		EventID    string          `json:"event_id"`
		RelayID    string          `json:"relay_id"`
		Payload    json.RawMessage `json:"payload"`
		ReceivedAt string          `json:"received_at"`
	}
	var evt Event
	if err := json.Unmarshal(msg.Data, &evt); err != nil {
		c.logger.Error("failed to parse message",
			slog.String("error", err.Error()))
		msg.Nak()
		return
	}
	c.logger.Debug("received event",
		slog.String("relay_id", evt.RelayID),
		slog.String("event_id", evt.EventID),
		slog.Int("payload_size", len(evt.Payload)))
	// Bridges NATS consumer to Worker Pool
	job := engine.Job{
		RelayID: evt.RelayID,
		EventID: evt.EventID,
		Payload: evt.Payload,
		MsgAck: func(success bool) {
			if success {
				msg.Ack()
				c.logger.Debug("acknowledged message", slog.String("relay_id", evt.RelayID),
					slog.String("event_id", evt.EventID))
			} else {
				msg.Nak()
				c.logger.Warn("nacked message (will retry)", slog.String("relay_id", evt.RelayID),
					slog.String("event_id", evt.EventID))
			}
		},
	}
	//Blocking send to channel - If the worker is full this will wait
	c.jobQueue <- job
}

func (c *Consumer) Stop() error {
	c.logger.Info("stopping NATS consumer")
	if c.sub != nil {
		// To process remaining messages and then close
		return c.sub.Drain()
	}
	return nil
}
