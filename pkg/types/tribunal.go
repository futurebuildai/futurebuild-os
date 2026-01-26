package types

// TribunalDecision represents a self-healing diagnostic decision.
// Used by the Tree Planting integration test for FutureShade's
// autonomous diagnosis and remediation capabilities.
type TribunalDecision struct {
	// FaultDiagnosis is the identified fault type (e.g., "CONFIG_DRIFT").
	FaultDiagnosis string `json:"fault_diagnosis"`

	// ConfidenceScore is the AI's confidence in the diagnosis (0.0-1.0).
	ConfidenceScore float64 `json:"confidence_score"`

	// Reasoning explains the diagnostic process and evidence.
	Reasoning string `json:"reasoning"`

	// ProposedAction is the recommended remediation action.
	ProposedAction ReformAction `json:"proposed_action"`
}

// ReformAction represents a remediation action proposed by the Tribunal.
type ReformAction struct {
	// Type is the category of action to take.
	Type ReformActionType `json:"type"`

	// Key is the configuration key or resource identifier to modify.
	Key string `json:"key"`

	// Value is the new value to set (can be any JSON-serializable type).
	Value interface{} `json:"value"`
}

// ReformActionType defines the category of remediation action.
type ReformActionType string

const (
	// ActionUpdateConfig updates an in-memory configuration value.
	ActionUpdateConfig ReformActionType = "UPDATE_CONFIG"

	// ActionClearCache clears a specified cache.
	ActionClearCache ReformActionType = "CLEAR_CACHE"

	// ActionRetry retries the failed operation.
	ActionRetry ReformActionType = "RETRY"

	// ActionNoOp indicates no action is needed.
	ActionNoOp ReformActionType = "NO_OP"
)

// IsValidActionType checks if a string is a valid ReformActionType.
// Safety: Only allows approved action types to prevent arbitrary operations.
func IsValidActionType(s string) bool {
	switch ReformActionType(s) {
	case ActionUpdateConfig, ActionClearCache, ActionRetry, ActionNoOp:
		return true
	default:
		return false
	}
}
