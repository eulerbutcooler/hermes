package httpreq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/engine"
)

var allowedMethods = map[string]bool{
	http.MethodGet:    true,
	http.MethodPost:   true,
	http.MethodPut:    true,
	http.MethodDelete: true,
	http.MethodPatch:  true,
}

type HTTPRequestOutput struct {
	StatusCode  int               `json:"status_code"`
	ContentType string            `json:"content_type,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	BodyJSON    any               `json:"body_json,omitempty"`
	BodyText    string            `json:"body_text,omitempty"`
	DurationMS  int64             `json:"duration_ms,omitempty"`
}

type Executor struct {
	client *http.Client
}

func New() *Executor {
	return &Executor{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (e *Executor) Execute(ctx context.Context, cfg map[string]any, payload []byte, _ map[string]engine.StepOutput) (json.RawMessage, error) {
	url, _ := cfg["url"].(string)
	if url == "" {
		return nil, fmt.Errorf("missing url in http_request config")
	}

	method, _ := cfg["method"].(string)
	method = strings.ToUpper(method)
	if method == "" {
		method = http.MethodPost
	}
	if !allowedMethods[method] {
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	var lastErr error
	for attempt := range 3 {
		var body io.Reader
		if method != http.MethodGet {
			bodyTemplate, _ := cfg["body_template"].(string)
			if bodyTemplate != "" {
				body = strings.NewReader(bodyTemplate)
			} else {
				body = bytes.NewReader(payload)
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return nil, fmt.Errorf("build request: %w", err)
		}

		if method != http.MethodGet {
			req.Header.Set("Content-Type", "application/json")
		}

		if headers, ok := cfg["headers"].(map[string]any); ok {
			for k, v := range headers {
				if s, ok := v.(string); ok {
					req.Header.Set(k, s)
				}
			}
		}

		start := time.Now()
		resp, doErr := e.client.Do(req)
		duration := time.Since(start)

		if doErr != nil {
			lastErr = doErr
			time.Sleep(time.Duration(300*(attempt+1)) * time.Millisecond)
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			output := HTTPRequestOutput{
				StatusCode:  resp.StatusCode,
				ContentType: resp.Header.Get("Content-Type"),
				Headers:     flattenHeaders(resp.Header),
				DurationMS:  duration.Milliseconds(),
			}

			if looksLikeJSON(output.ContentType, respBody) {
				var parsed any
				if err := json.Unmarshal(respBody, &parsed); err == nil {
					output.BodyJSON = parsed
				} else {
					output.BodyText = string(respBody)
				}
			} else if len(respBody) > 0 {
				output.BodyText = string(respBody)
			}

			encoded, err := json.Marshal(output)
			if err != nil {
				return nil, fmt.Errorf("marshal http_request output: %w", err)
			}
			return encoded, nil
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("http_request returned %d", resp.StatusCode)
			time.Sleep(time.Duration(300*(attempt+1)) * time.Millisecond)
			continue
		}

		return nil, fmt.Errorf("http_request returned non-retryable status %d", resp.StatusCode)
	}

	return nil, fmt.Errorf("http_request failed after retries: %w", lastErr)
}

func flattenHeaders(headers http.Header) map[string]string {
	if len(headers) == 0 {
		return nil
	}

	out := make(map[string]string, len(headers))
	for k, v := range headers {
		out[k] = strings.Join(v, ", ")
	}
	return out
}

func looksLikeJSON(contentType string, body []byte) bool {
	if strings.Contains(strings.ToLower(contentType), "application/json") {
		return true
	}

	trimmed := bytes.TrimSpace(body)
	return len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[')
}
