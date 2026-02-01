package models

import (
	"time"
)

// OnboardRequest is the incoming payload from the frontend wizard.
type OnboardRequest struct {
	SessionID    string                 `json:"session_id"`
	Message      string                 `json:"message,omitempty"`
	DocumentURL  string                 `json:"document_url,omitempty"`
	CurrentState map[string]interface{} `json:"current_state"`
}

// OnboardResponse is returned to the frontend with extracted values.
type OnboardResponse struct {
	SessionID          string             `json:"session_id"`
	Reply              string             `json:"reply"`
	ExtractedValues    map[string]any     `json:"extracted_values"`
	ConfidenceScores   map[string]float64 `json:"confidence_scores"`
	ClarifyingQuestion string             `json:"clarifying_question,omitempty"`
	ReadyToCreate      bool               `json:"ready_to_create"`
	NextPriorityField  string             `json:"next_priority_field,omitempty"`
}

// OnboardingSession persists wizard state (optional for MVP).
type OnboardingSession struct {
	ID                string                 `json:"id"`
	TenantID          string                 `json:"tenant_id"`
	UserID            string                 `json:"user_id"`
	FormState         map[string]interface{} `json:"form_state"`
	ExtractionHistory []ExtractionResult     `json:"extraction_history"`
	Status            string                 `json:"status"` // in_progress, completed, abandoned
	CreatedAt         time.Time              `json:"created_at"`
}

// ExtractionResult logs what was extracted from a document.
type ExtractionResult struct {
	DocumentURL string             `json:"document_url"`
	ExtractedAt time.Time          `json:"extracted_at"`
	Values      map[string]any     `json:"values"`
	Confidence  map[string]float64 `json:"confidence"`
}

// PhysicsFieldPriority defines extraction priority.
// P0 = critical, P1 = important, P2 = helpful
type PhysicsFieldPriority struct {
	Field    string
	Priority int // 0, 1, or 2
	Question string
}

// GetPriorityFields returns the ordered list of physics fields.
func GetPriorityFields() []PhysicsFieldPriority {
	return []PhysicsFieldPriority{
		{"name", 0, "What would you like to call this project?"},
		{"address", 0, "Where is the project located?"},
		{"gsf", 0, "What's the approximate square footage?"},
		{"foundation_type", 0, "What type of foundation? Slab, crawlspace, or basement?"},
		{"stories", 1, "Is this a single-story or multi-story home?"},
		{"topography", 1, "Is the lot flat, moderately sloped, or steeply sloped?"},
		{"soil_conditions", 1, "Any special soil conditions like rock or clay?"},
		{"bedrooms", 2, "How many bedrooms?"},
		{"bathrooms", 2, "How many bathrooms?"},
		{"supply_chain_volatility", 2, "Any supply chain concerns for this project?"},
	}
}
