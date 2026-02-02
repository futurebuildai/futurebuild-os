package types

import (
	"github.com/google/uuid"
)

// Forecast represents weather integration data.
// See API_AND_TYPES_SPEC.md Section 2.1
type Forecast struct {
	Date                     string  `json:"date"`
	HighTempC                float64 `json:"high_temp_c"`
	LowTempC                 float64 `json:"low_temp_c"`
	PrecipitationMM          float64 `json:"precipitation_mm"`
	PrecipitationProbability float64 `json:"precipitation_probability"`
	Conditions               string  `json:"conditions"`
}

// Contact represents a shared contact model.
// See API_AND_TYPES_SPEC.md Section 4.1
type Contact struct {
	ID                uuid.UUID         `json:"id"`
	Name              string            `json:"name"`
	Company           string            `json:"company"`
	Phone             string            `json:"phone"`
	Email             string            `json:"email"`
	Role              UserRole          `json:"role"`
	ContactPreference ContactPreference `json:"contact_preference,omitempty"` // See DATA_SPINE_SPEC.md Section 2.3
}

// InvoiceExtraction represents the output of document analysis.
// See API_AND_TYPES_SPEC.md Section 3.1
// MONETARY PRECISION: All amounts as int64 cents to prevent IEEE 754 float drift.
type InvoiceExtraction struct {
	Vendor           string                  `json:"vendor"`
	Date             string                  `json:"date"`
	InvoiceNumber    string                  `json:"invoice_number"`
	TotalAmountCents int64                   `json:"total_amount_cents,string"`
	LineItems        []InvoiceExtractionItem `json:"line_items"`
	SuggestedWBSCode string                  `json:"suggested_wbs_code"`
	Confidence       float64                 `json:"confidence"` // Confidence remains float (0.0-1.0)
}

// InvoiceExtractionItem represents a single line item in an invoice.
// MONETARY PRECISION: UnitPrice and Total as int64 cents.
type InvoiceExtractionItem struct {
	Description    string  `json:"description"`
	Quantity       float64 `json:"quantity"` // Kept as float - quantities can be fractional
	UnitPriceCents int64   `json:"unit_price_cents,string"`
	TotalCents     int64   `json:"total_cents,string"`
}

// GanttData represents the project schedule for the Gantt view.
// See API_AND_TYPES_SPEC.md Section 3.2
type GanttData struct {
	ProjectID        uuid.UUID         `json:"project_id"`
	CalculatedAt     string            `json:"calculated_at"`
	ProjectedEndDate string            `json:"projected_end_date"`
	CriticalPath     []string          `json:"critical_path"`
	Tasks            []GanttTask       `json:"tasks"`
	Dependencies     []GanttDependency `json:"dependencies,omitempty"` // Step 89: Dependency edges for SVG arrows
}

// GanttTask represents an individual task in the Gantt data.
type GanttTask struct {
	WBSCode      string     `json:"wbs_code"`
	Name         string     `json:"name"`
	Status       TaskStatus `json:"status"`
	EarlyStart   string     `json:"early_start"`
	EarlyFinish  string     `json:"early_finish"`
	DurationDays float64    `json:"duration_days"`
	IsCritical   bool       `json:"is_critical"`
}

// GanttDependency represents a directed edge between two tasks (Finish-to-Start).
// See STEP_89_DEPENDENCY_ARROWS.md Section 1.2
type GanttDependency struct {
	From string `json:"from"` // Predecessor WBS code
	To   string `json:"to"`   // Successor WBS code
}

// AuthRequest is the payload for requesting a portal magic link.
type AuthRequest struct {
	Email string `json:"email"`
}

// AuthResponse is the generic response for portal auth requests.
type AuthResponse struct {
	Message string `json:"message"`
}

// User represents an internal user.
// See DATA_SPINE_SPEC.md Section 2.2
type User struct {
	ID         uuid.UUID `json:"id"`
	OrgID      uuid.UUID `json:"org_id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	Role       UserRole  `json:"role"`
	ExternalID string    `json:"external_id,omitempty"` // Clerk user ID (Step 80)
	CreatedAt  string    `json:"created_at"`
}

// ToolCall represents a single tool invocation from the AI model.
// See API_AND_TYPES_SPEC.md Section 4.4 (Chat Domain)
// CTO-002 Remediation: Typed struct instead of interface{}
type ToolCall struct {
	ID       string                 `json:"id"`       // Unique ID for the tool call
	Name     string                 `json:"name"`     // Tool function name
	Args     map[string]interface{} `json:"args"`     // Arguments passed to the tool
	Response string                 `json:"response"` // Tool output (filled after execution)
}

// ChatMessage represents a single message in an agent session.
// See API_AND_TYPES_SPEC.md Section 4.4
type ChatMessage struct {
	ID        uuid.UUID  `json:"id"`         // UUID string
	ProjectID uuid.UUID  `json:"project_id"` // UUID string
	UserID    uuid.UUID  `json:"user_id"`    // UUID string
	Role      ChatRole   `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"` // CTO-002: Typed struct
	CreatedAt string     `json:"created_at"`           // ISO-8601 Timestamp
}

// DynamicComponent represents a recursive UI element.
// See API_AND_TYPES_SPEC.md Section 4.5
type DynamicComponent struct {
	Type     DynamicComponentType   `json:"type"`
	Props    map[string]interface{} `json:"props"`
	Children []DynamicComponent     `json:"children,omitempty"`
	ActionID string                 `json:"action_id,omitempty"`
}

// DynamicUIArtifact represents a full UI definition served by the agent.
// See API_AND_TYPES_SPEC.md Section 4.5
type DynamicUIArtifact struct {
	Title     string           `json:"title"`
	Root      DynamicComponent `json:"root"`
	SubmitURL string           `json:"submit_url,omitempty"`
}
