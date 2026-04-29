package engine

import (
	"context"
	"encoding/json"
	"testing"
)

// Implements ActionExecutor for testing
type fakeExecutor struct {
	called bool
	err    error
	output json.RawMessage
}

func (f *fakeExecutor) Execute(ctx context.Context, config map[string]any, payload []byte, prevOutputs map[string]StepOutput) (json.RawMessage, error) {
	f.called = true
	return f.output, f.err
}

func TestRegistery_RegisterAndGet(t *testing.T) {
	reg := NewRegistry()
	exec := &fakeExecutor{}

	reg.Register("test_action", exec)
	got, err := reg.Get("test_action")
	if err != nil {
		t.Fatalf("'Get' returned error: %v", err)
	}
	if _, err := got.Execute(context.Background(), nil, nil, nil); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if !exec.called {
		t.Error("expected executor to be called")
	}
}

// Verifies that requesting an unregistered type returns an error
func TestRegister_GetUnknown(t *testing.T) {
	reg := NewRegistry()

	_, err := reg.Get("bruh")
	if err == nil {
		t.Error("expected error for unknown action type, got nil")
	}
}

// Verifies the Count() helper
func TestRegistry_Count(t *testing.T) {
	reg := NewRegistry()

	if reg.Count() != 0 {
		t.Errorf("expected 0, got %d", reg.Count())
	}

	reg.Register("a", &fakeExecutor{})
	reg.Register("b", &fakeExecutor{})

	if reg.Count() != 2 {
		t.Errorf("expected 2, got %d", reg.Count())
	}
}

// Verifies that Types() returns sorted names.
func TestRegistry_Types(t *testing.T) {
	reg := NewRegistry()
	reg.Register("apple", &fakeExecutor{})
	reg.Register("zango", &fakeExecutor{})

	types := reg.Types()
	if len(types) != 2 {
		t.Fatalf("expected 2 types, got %d", len(types))
	}
	// Must be sorted alphabetically.
	if types[0] != "apple" || types[1] != "zango" {
		t.Errorf("expected [apple, zango], got %v", types)
	}
}

// Verifies the Has() helper.
func TestRegistry_Has(t *testing.T) {
	reg := NewRegistry()
	reg.Register("exists", &fakeExecutor{})

	if !reg.Has("exists") {
		t.Error("Has('exists') = false, want true")
	}
	if reg.Has("nope") {
		t.Error("Has('nope') = true, want false")
	}
}
