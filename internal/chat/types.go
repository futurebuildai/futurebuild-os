package chat

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/colton/futurebuild/pkg/types"
)

// --- Artifact Models (Rich UI Support) ---
// See PRODUCTION_PLAN.md Step 44 (Internal Artifact Mapping)

// ArtifactType defines the visual component type returned by the Orchestrator.
type ArtifactType string

const (
	// ArtifactTypeScheduleView triggers a Gantt/schedule visualization.
	ArtifactTypeScheduleView ArtifactType = "schedule_view"

	// ArtifactTypeInvoiceReview triggers an invoice processing UI.
	ArtifactTypeInvoiceReview ArtifactType = "invoice_review"

	// ArtifactTypeDailyBriefing triggers a daily summary view.
	ArtifactTypeDailyBriefing ArtifactType = "daily_briefing"
)

// Artifact represents a structured data payload for Rich UI rendering.
// Returned alongside text responses for visualization components.
// See PRODUCTION_PLAN.md Step 44
type Artifact struct {
	Type  ArtifactType    `json:"type"`
	Data  json.RawMessage `json:"data"`
	Title string          `json:"title"`
}

// ChatRequest defines the contract for an inbound user message.
// detailed in PRODUCTION_PLAN.md Step 43.
// NOTE: UserID is deliberately omitted here. It is injected from the
// authenticated Request Context (JWT) by the Orchestrator to prevent spoofing.
type ChatRequest struct {
	ProjectID uuid.UUID `json:"project_id"`
	Message   string    `json:"message"`
}

// ChatResponse defines the structured reply from the Orchestrator.
// See PRODUCTION_PLAN.md Step 44 (Artifact field for Rich UI)
type ChatResponse struct {
	Reply    string       `json:"reply"`
	Intent   types.Intent `json:"intent"`
	Artifact *Artifact    `json:"artifact,omitempty"`
}
