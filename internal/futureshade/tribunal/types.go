// Package tribunal contains types and logic for The Tribunal consensus system.
// See specs/SHADOW_VIEWER_specs.md Section 3.1 and 4.1
package tribunal

import (
	"errors"
	"time"

	"github.com/colton/futurebuild/pkg/ai"
	"github.com/google/uuid"
)

// ErrNotFound is returned when a decision is not found.
var ErrNotFound = errors.New("decision not found")

// DecisionStatus represents the outcome of a Tribunal decision.
// Maps to tribunal_decision_status enum in database.
type DecisionStatus string

const (
	DecisionPending  DecisionStatus = "pending"
	DecisionApproved DecisionStatus = "APPROVED"
	DecisionRejected DecisionStatus = "REJECTED"
	DecisionConflict DecisionStatus = "CONFLICT"
)

// VoteType represents an individual model's vote.
// Maps to tribunal_vote_type enum in database.
type VoteType string

const (
	VoteYea     VoteType = "YEA"
	VoteNay     VoteType = "NAY"
	VoteAbstain VoteType = "ABSTAIN"
)

// ModelVote represents a single model's vote in a Tribunal decision.
type ModelVote struct {
	ID         uuid.UUID `json:"id,omitempty"`
	DecisionID uuid.UUID `json:"decision_id,omitempty"`
	ModelName  string    `json:"model"`
	Vote       VoteType  `json:"vote"`
	Reasoning  string    `json:"reasoning"`
	LatencyMs  int       `json:"latency_ms"`
	TokenCount int       `json:"token_count,omitempty"`
	CostUSD    float64   `json:"cost_usd"`
}

// DecisionSummary is the list view response for tribunal decisions.
// See SHADOW_VIEWER_specs.md Section 3.1 GET /api/v1/tribunal/decisions
type DecisionSummary struct {
	ID              uuid.UUID      `json:"id"`
	CaseID          string         `json:"case_id"`
	Status          DecisionStatus `json:"status"`
	Context         string         `json:"context"`
	Timestamp       time.Time      `json:"timestamp"`
	ModelsConsulted []string       `json:"models_consulted"`
}

// DecisionDetail is the full detail view response including individual model votes.
// See SHADOW_VIEWER_specs.md Section 3.1 GET /api/v1/tribunal/decisions/{id}
type DecisionDetail struct {
	ID             uuid.UUID      `json:"id"`
	CaseID         string         `json:"case_id"`
	Status         DecisionStatus `json:"status"`
	Context        string         `json:"context"`
	ConsensusScore float64        `json:"consensus_score"`
	Votes          []ModelVote    `json:"votes"`
	PolicyLinks    []string       `json:"policy_links,omitempty"`
	Timestamp      time.Time      `json:"timestamp"`
}

// ListDecisionsFilter holds query parameters for filtering decisions.
type ListDecisionsFilter struct {
	Limit     int            `json:"limit"`
	Offset    int            `json:"offset"`
	Status    DecisionStatus `json:"status,omitempty"`
	Model     string         `json:"model,omitempty"`
	StartDate *time.Time     `json:"start_date,omitempty"`
	EndDate   *time.Time     `json:"end_date,omitempty"`
	Search    string         `json:"search,omitempty"`
}

// ListDecisionsResponse is the paginated response for the list endpoint.
type ListDecisionsResponse struct {
	Decisions []DecisionSummary `json:"decisions"`
	Total     int               `json:"total"`
	HasMore   bool              `json:"has_more"`
}

// TribunalRequest represents a request for a Tribunal decision.
type TribunalRequest struct {
	CaseID  string `json:"case_id"`
	Intent  string `json:"intent"`  // The standardized intent or problem description
	Context string `json:"context"` // Additional context (file snapshots, diffs)
}

// TribunalResponse represents the final consensus decision.
type TribunalResponse struct {
	DecisionID     uuid.UUID      `json:"decision_id"`
	Status         DecisionStatus `json:"status"`
	ConsensusScore float64        `json:"consensus_score"`
	Summary        string         `json:"summary"` // Synthesized reasoning
	Plan           string         `json:"plan"`    // The recommended action/plan
}

// Jury represents the panel of models evaluating the case.
type Jury struct {
	Coordinator ai.Client // Gemini 3 Flash
	Architect   ai.Client // Claude Opus
	Historian   ai.Client // Gemini Code Assist
}
