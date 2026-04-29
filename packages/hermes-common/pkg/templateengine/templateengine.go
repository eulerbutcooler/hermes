package templateengine

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type StepOutput struct {
	ActionType string          `json:"action_type"`
	NodeID     string          `json:"node_id"`
	Output     json.RawMessage `json:"output"`
}

// Matches {{...}} pattern
var templatePattern = regexp.MustCompile(`\{\{\s*(.+?)\s*\}\}`)

// Replaces any string values that contain {{ epxr }}
func Resolve(config map[string]any, payload []byte, steps map[string]StepOutput) map[string]any {
	out := make(map[string]any, len(config))
	for k, v := range config {
		switch val := v.(type) {
		case string:
			out[k] = resolveString(val, payload, steps)
		default:
			out[k] = v
		}
	}
	return out
}

func resolveString(s string, payload []byte, steps map[string]StepOutput) string {
	return templatePattern.ReplaceAllStringFunc(s, func(match string) string {
		inner := templatePattern.FindStringSubmatch(match)
		if len(inner) < 2 {
			return match
		}
		expr := strings.TrimSpace(inner[1])
		val, err := evaluate(expr, payload, steps)
		if err != nil {
			return match
		}
		return val
	})
}

func evaluate(expr string, payload []byte, steps map[string]StepOutput) (string, error) {
	parts := strings.SplitN(expr, ".", 2)
	root := parts[0]

	switch {
	case root == "payload":
		if len(parts) == 1 {
			return string(payload), nil
		}
		return drillJSON(payload, parts[1])

	case strings.HasPrefix(root, "steps['"):
		nodeID := strings.TrimPrefix(root, "steps['")
		nodeID = strings.TrimSuffix(nodeID, "']")
		step, ok := steps[nodeID]

		if !ok {
			return "", fmt.Errorf("step %q not found in outputs", nodeID)
		}

		if len(parts) == 1 {
			return string(step.Output), nil
		}
		rest := parts[1]
		rest = strings.TrimPrefix(rest, "output")
		rest = strings.TrimPrefix(rest, ".")
		if rest == "" {
			return string(step.Output), nil
		}
		return drillJSON(step.Output, rest)

	default:
		return "", fmt.Errorf("unknown template root: %s", root)
	}
}

func parseStepIndex(s string) (int, error) {
	s = strings.TrimPrefix(s, "steps[")
	s = strings.TrimSuffix(s, "]")
	return strconv.Atoi(s)
}

func drillJSON(raw []byte, path string) (string, error) {
	if len(raw) == 0 {
		return "", fmt.Errorf("empty JSON")
	}
	var obj any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	segments := strings.Split(path, ".")
	current := obj
	for _, seg := range segments {
		if before, after, ok := strings.Cut(seg, "["); ok {
			key := before
			idxStr := strings.TrimSuffix(after, "]")
			idx, err := strconv.Atoi(idxStr)
			if err != nil {
				return "", fmt.Errorf("invalid array index %q in %q", idxStr, seg)
			}
			if key != "" {
				m, ok := current.(map[string]any)
				if !ok {
					return "", fmt.Errorf("expected object at %q, got %T", key, current)
				}
				current, ok = m[key]
				if !ok {
					return "", fmt.Errorf("key %q not found", key)
				}
			}
			arr, ok := current.([]any)
			if !ok {
				return "", fmt.Errorf("expected array at %q, got %T", seg, current)
			}
			if idx < 0 || idx >= len(arr) {
				return "", fmt.Errorf("array index %d out of range (len %d)", idx, len(arr))
			}
			current = arr[idx]
		} else {
			m, ok := current.(map[string]any)
			if !ok {
				return "", fmt.Errorf("cannot drill into non-object at %q", seg)
			}
			current, ok = m[seg]
			if !ok {
				return "", fmt.Errorf("key %q not found", seg)
			}
		}
	}

	switch v := current.(type) {
	case string:
		return v, nil
	default:
		b, _ := json.Marshal(v)
		return string(b), nil
	}
}

func EvaluateExpression(expr string, payload []byte, steps map[string]StepOutput) (string, error) {
	return evaluate(expr, payload, steps)
}

func EvaluateCondition(cond map[string]any, payload []byte, steps map[string]StepOutput) bool {
	if cond == nil {
		return true
	}

	fieldRaw, ok := cond["field"].(string)
	if !ok || strings.TrimSpace(fieldRaw) == "" {
		return true
	}

	operator, _ := cond["operator"].(string)
	valueRaw, _ := cond["value"].(string)

	fieldExpr := strings.TrimPrefix(fieldRaw, "{{")
	fieldExpr = strings.TrimSuffix(fieldExpr, "}}")
	fieldExpr = strings.TrimSpace(fieldExpr)

	actualValue, err := evaluate(fieldExpr, payload, steps)
	if err != nil {
		actualValue = ""
	}

	switch operator {
	case "==":
		return actualValue == valueRaw
	case "!=":
		return actualValue != valueRaw
	case ">":
		actualF, err1 := strconv.ParseFloat(actualValue, 64)
		expectedF, err2 := strconv.ParseFloat(valueRaw, 64)
		if err1 == nil && err2 == nil {
			return actualF > expectedF
		}
		return actualValue > valueRaw
	case ">=":
		actualF, err1 := strconv.ParseFloat(actualValue, 64)
		expectedF, err2 := strconv.ParseFloat(valueRaw, 64)
		if err1 == nil && err2 == nil {
			return actualF >= expectedF
		}
		return actualValue >= valueRaw
	case "<":
		actualF, err1 := strconv.ParseFloat(actualValue, 64)
		expectedF, err2 := strconv.ParseFloat(valueRaw, 64)
		if err1 == nil && err2 == nil {
			return actualF < expectedF
		}
		return actualValue < valueRaw
	case "<=":
		actualF, err1 := strconv.ParseFloat(actualValue, 64)
		expectedF, err2 := strconv.ParseFloat(valueRaw, 64)
		if err1 == nil && err2 == nil {
			return actualF <= expectedF
		}
		return actualValue <= valueRaw
	case "contains":
		return strings.Contains(actualValue, valueRaw)
	case "not_contains":
		return !strings.Contains(actualValue, valueRaw)
	default:
		return actualValue == valueRaw
	}
}
