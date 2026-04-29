package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/engine"
)

type Sender struct {
	client *http.Client
}

type DiscordSendOutput struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	DurationMS int64  `json:"duration_ms,omitempty"`
}

func New() *Sender {
	return &Sender{
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (d *Sender) Execute(ctx context.Context, config map[string]any, payload []byte, _ map[string]engine.StepOutput) (json.RawMessage, error) {
	url, _ := config["webhook_url"].(string)
	if url == "" {
		return nil, fmt.Errorf("missing webhook_url in discord config")
	}

	messageTemplate, _ := config["message_template"].(string)
	var content string
	if messageTemplate != "" {
		content = messageTemplate
	} else {
		content = fmt.Sprintf("Relay Triggered\n```json\n%s\n```", string(payload))
	}

	msg := map[string]string{"content": content}
	jsonBody, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal discord body: %w", err)
	}

	var lastErr error
	for attempt := range 3 {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
		if reqErr != nil {
			return nil, fmt.Errorf("build request: %w", reqErr)
		}
		req.Header.Set("Content-Type", "application/json")

		start := time.Now()
		resp, doErr := d.client.Do(req)
		duration := time.Since(start)

		if doErr != nil {
			lastErr = doErr
			time.Sleep(time.Duration(200*(attempt+1)) * time.Millisecond)
			continue
		}

		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			output, err := json.Marshal(DiscordSendOutput{
				Status:     "sent",
				StatusCode: resp.StatusCode,
				DurationMS: duration.Milliseconds(),
			})
			if err != nil {
				return nil, fmt.Errorf("marshal discord output: %w", err)
			}
			return output, nil
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("discord returned %d", resp.StatusCode)
			time.Sleep(time.Duration(200*(attempt+1)) * time.Millisecond)
			continue
		}

		return nil, fmt.Errorf("discord returned non-retryable status %d", resp.StatusCode)
	}

	return nil, fmt.Errorf("discord send failed after retries: %w", lastErr)
}
