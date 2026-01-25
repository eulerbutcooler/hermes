package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Config struct {
	WebhookURL      string
	MessageTemplate string
}

type Sender struct {
	client *http.Client
}

func New() *Sender {
	return &Sender{
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *Sender) Execute(ctx context.Context, cfg map[string]any, payload []byte) error {
	webhookURL, _ := cfg["webhook_url"].(string)
	template, _ := cfg["message_template"].(string)

	if webhookURL == "" {
		return fmt.Errorf("missing webhook_url in slack action config")
	}
	var text string
	if template != "" {
		text = template
	} else {
		text = fmt.Sprintf("Payload:\n```json\n%s\n```", string(payload))
	}
	bodyMap := map[string]any{
		"text": text,
	}

	bodyJSON, err := json.Marshal(bodyMap)
	if err != nil {
		return fmt.Errorf("marshal slack body: %w", err)
	}

	var lastErr error
	for attempt := range 3 {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewBuffer(bodyJSON))
		if reqErr != nil {
			return fmt.Errorf("build request: %w", reqErr)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, doErr := s.client.Do(req)
		if doErr != nil {
			lastErr = doErr
		} else {
			resp.Body.Close()
			if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
				lastErr = fmt.Errorf("slack returned %d", resp.StatusCode)
			} else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return nil
			} else {
				return fmt.Errorf("slack returned non-retryable status %d", resp.StatusCode)
			}
		}
		time.Sleep(time.Duration(200*(attempt+1)) * time.Millisecond)
	}
	return fmt.Errorf("slack send failed after retries: %w", lastErr)
}
