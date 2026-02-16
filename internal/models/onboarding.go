package models

import (
	"time"
)

// OnboardRequest is the incoming payload from the frontend wizard.
// Supports both JSON (text/document_url) and multipart/form-data (inline file upload).
type OnboardRequest struct {
	SessionID    string                 `json:"session_id"`
	Message      string                 `json:"message,omitempty"`
	DocumentURL  string                 `json:"document_url,omitempty"`
	CurrentState map[string]interface{} `json:"current_state"`

	// Inline file data (populated by multipart handler, not JSON).
	// Step 77: Magic Upload Trigger - direct file upload path.
	DocumentData        []byte `json:"-"`
	DocumentContentType string `json:"-"`
	DocumentFileName    string `json:"-"`
}

// LongLeadItem represents a material or equipment with significant lead time.
// These items affect schedule generation and need early procurement.
type LongLeadItem struct {
	Name               string `json:"name"`
	Brand              string `json:"brand,omitempty"`
	Model              string `json:"model,omitempty"`
	Category           string `json:"category"` // windows, doors, hvac, appliances, millwork, finishes
	EstimatedLeadWeeks int    `json:"estimated_lead_weeks"`
	WBSCode            string `json:"wbs_code,omitempty"` // Affected WBS code (e.g., "8.1" for windows)
	Notes              string `json:"notes,omitempty"`
}

// OnboardResponse is returned to the frontend with extracted values.
type OnboardResponse struct {
	SessionID          string             `json:"session_id"`
	Reply              string             `json:"reply"`
	ExtractedValues    map[string]any     `json:"extracted_values"`
	ConfidenceScores   map[string]float64 `json:"confidence_scores"`
	LongLeadItems      []LongLeadItem     `json:"long_lead_items,omitempty"`
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
	DocumentURL   string             `json:"document_url"`
	ExtractedAt   time.Time          `json:"extracted_at"`
	Values        map[string]any     `json:"values"`
	Confidence    map[string]float64 `json:"confidence"`
	LongLeadItems []LongLeadItem     `json:"long_lead_items,omitempty"`
}

// ConfidenceReport provides per-field confidence data so the frontend
// can badge low-confidence fields for user verification.
// Sprint 2.1 Task 2.1.2: Interrogator Gate — Onboarding Intelligence.
type ConfidenceReport struct {
	OverallConfidence  float64            `json:"overall_confidence"`
	FieldConfidences   map[string]float64 `json:"field_confidences"`
	Warnings           []string           `json:"warnings"`
	SuggestedQuestions []string           `json:"suggested_questions"`
}

// VisionExtractionResponse is returned by the standalone POST /api/v1/vision/extract endpoint.
// It wraps extracted values with a full ConfidenceReport for the frontend.
type VisionExtractionResponse struct {
	ExtractedValues  map[string]any   `json:"extracted_values"`
	ConfidenceReport ConfidenceReport `json:"confidence_report"`
	LongLeadItems    []LongLeadItem   `json:"long_lead_items,omitempty"`
	RawText          string           `json:"raw_text,omitempty"`
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
		{"start_date", 0, "When did your permit get issued, or when do you plan to break ground?"},
		{"square_footage", 0, "What's the approximate square footage?"},
		{"foundation_type", 0, "What type of foundation? Slab, crawlspace, or basement?"},
		{"stories", 1, "Is this a single-story or multi-story home?"},
		{"topography", 1, "Is the lot flat, sloped, or hillside?"},
		{"soil_conditions", 1, "Any special soil conditions? Normal, rocky, clay, or sandy?"},
		{"bedrooms", 2, "How many bedrooms?"},
		{"bathrooms", 2, "How many bathrooms?"},
		{"holidays", 2, "Any holidays or planned breaks I should account for?"},
	}
}

// KnownBrandLeadTimes returns estimated lead times in weeks for known brands.
// Used to calculate procurement constraints for schedule generation.
func KnownBrandLeadTimes() map[string]int {
	return map[string]int{
		// Windows
		"marvin":          12,
		"marvin ultimate": 14,
		"andersen":        10,
		"andersen e":      12,
		"pella":           10,
		"pella reserve":   12,
		"milgard":         5,

		// Appliances
		"sub-zero":   10,
		"subzero":    10,
		"wolf":       10,
		"viking":     8,
		"la cornue":  20,
		"thermador":  8,
		"miele":      8,
		"gaggenau":   10,

		// HVAC
		"geothermal": 8,
		"carrier":    4,
		"trane":      4,
		"lennox":     4,

		// Doors
		"simpson":    6,
		"therma-tru": 6,
	}
}
