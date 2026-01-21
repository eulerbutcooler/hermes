package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DiscordSender struct {
	client *http.Client
}

func New() *DiscordSender {
	return &DiscordSender{
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (d *DiscordSender) Execute(ctx context.Context, config map[string]any, payload []byte) error {
	url, ok := config["webhook_url"].(string)
	if !ok || url == "" {
		return fmt.Errorf("Missing webhook_url in relay config")
	}
	msg := map[string]string{
		"content": fmt.Sprintf("Relay Trigerred\n```json\n%s\n```", string(payload)),
	}
	jsonBody, _ := json.Marshal(msg)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 400 {
		return fmt.Errorf("Discord API error: %d", resp.StatusCode)
	}
	return nil
}
