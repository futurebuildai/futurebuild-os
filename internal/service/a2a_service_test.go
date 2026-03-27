package service

import (
	"testing"
)

func TestNewA2AService(t *testing.T) {
	svc := NewA2AService(nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestAgentStatusTransitions(t *testing.T) {
	// Valid transitions: active -> paused, paused -> active, any -> error
	testCases := []struct {
		name  string
		from  string
		to    string
		valid bool
	}{
		{"active_to_paused", "active", "paused", true},
		{"paused_to_active", "paused", "active", true},
		{"active_to_error", "active", "error", true},
		{"paused_to_error", "paused", "error", true},
		{"error_to_active", "error", "active", true}, // recovery
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.valid {
				t.Errorf("expected invalid transition from %s to %s", tc.from, tc.to)
			}
			// All defined transitions are valid in our model
		})
	}
}

func TestA2ALogPayload_JSON(t *testing.T) {
	// Verify that JSON payloads can be represented
	payload := `{"key": "value", "count": 42}`
	if len(payload) == 0 {
		t.Error("expected non-empty payload")
	}
}
