package chat

import (
	"github.com/google/uuid"

	"github.com/colton/futurebuild/pkg/types"
)

// ChatRequest defines the contract for an inbound user message.
// detailed in PRODUCTION_PLAN.md Step 43.
// NOTE: UserID is deliberately omitted here. It is injected from the
// authenticated Request Context (JWT) by the Orchestrator to prevent spoofing.
type ChatRequest struct {
	ProjectID uuid.UUID `json:"project_id"`
	Message   string    `json:"message"`
}

// ChatResponse defines the structured reply from the Orchestrator.
type ChatResponse struct {
	Reply  string       `json:"reply"`
	Intent types.Intent `json:"intent"`
}
