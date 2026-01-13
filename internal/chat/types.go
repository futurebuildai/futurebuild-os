package chat

import (
	"github.com/google/uuid"
)

// Intent is a string type alias for strict intent classification.
// See API_AND_TYPES_SPEC.md Section 4.3 (Dynamic UI)
type Intent string

const (
	// IntentProcessInvoice triggers the invoice extraction and verification flow.
	// See BACKEND_SCOPE.md Section 3.5 (Action Engine)
	IntentProcessInvoice Intent = "PROCESS_INVOICE"

	// IntentUnknown is the fallback when no intent is classified.
	IntentUnknown Intent = "UNKNOWN"
)

// ChatRequest defines the contract for an inbound user message.
// See PRODUCTION_PLAN.md Step 43
type ChatRequest struct {
	ProjectID uuid.UUID `json:"project_id"`
	Message   string    `json:"message"`
}

// ChatResponse defines the structured reply from the Orchestrator.
type ChatResponse struct {
	Reply  string `json:"reply"`
	Intent Intent `json:"intent"`
}
