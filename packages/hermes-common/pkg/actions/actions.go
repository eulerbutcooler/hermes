package actions

import (
	"fmt"
	"sort"
	"strings"
)

type FieldRule struct {
	// When Alternatives is empty, Name itself is the required key.
	Name string

	// At least one of these must be present
	// Example: ["webhook_url", "webhook_url_ref"]
	Alternatives []string
}

type ActionSchema struct {
	Fields []FieldRule

	// Extra is an optional, action‑specific validation hook.
	Extra func(config map[string]any) error
}

var Registry = map[string]ActionSchema{
	"debug_log": {
		Fields: nil,
	},
	"discord_send": {
		Fields: []FieldRule{
			{Name: "webhook_url", Alternatives: []string{"webhook_url", "webhook_url_ref"}},
		},
	},
	"slack_send": {
		Fields: []FieldRule{
			{Name: "webhook_url", Alternatives: []string{"webhook_url", "webhook_url_ref"}},
		},
	},
	"http_request": {
		Fields: []FieldRule{
			{Name: "url", Alternatives: []string{"url", "url_ref"}},
		},
		Extra: validateHTTPRequest,
	},
	"email_send": {
		Fields: []FieldRule{
			{Name: "connection_id", Alternatives: []string{"connection_id", "connection_id_ref"}},
			{Name: "to"},
		},
	},
	"condition": {
		Fields: nil,
	},
}

func IsValidType(actionType string) bool {
	_, ok := Registry[actionType]
	return ok
}

func Types() []string {
	out := make([]string, 0, len(Registry))
	for k := range Registry {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func ValidateConfig(actionType string, config map[string]any) error {
	schema, ok := Registry[actionType]
	if !ok {
		return fmt.Errorf("unknown action type: %s", actionType)
	}

	for _, rule := range schema.Fields {
		keys := rule.Alternatives
		if len(keys) == 0 {
			keys = []string{rule.Name}
		}
		if !hasOneOf(config, keys) {
			return fmt.Errorf("%s: requires one of [%s]", actionType, strings.Join(keys, ", "))
		}
	}

	if schema.Extra != nil {
		if err := schema.Extra(config); err != nil {
			return err
		}
	}

	return nil
}

func hasOneOf(cfg map[string]any, keys []string) bool {
	for _, key := range keys {
		if v, ok := cfg[key].(string); ok && strings.TrimSpace(v) != "" {
			return true
		}
	}
	return false
}

var allowedMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true,
}

func validateHTTPRequest(config map[string]any) error {
	if h, exists := config["headers"]; exists {
		if _, ok := h.(map[string]any); !ok {
			return fmt.Errorf("http_request: headers must be a JSON object")
		}
	}
	if m, exists := config["method"]; exists {
		method, _ := m.(string)
		if !allowedMethods[strings.ToUpper(method)] {
			return fmt.Errorf("http_request: method must be one of: GET, POST, PUT, PATCH, DELETE")
		}
	}
	return nil
}
