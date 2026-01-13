package types

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
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Company           string            `json:"company"`
	Phone             string            `json:"phone"`
	Email             string            `json:"email"`
	Role              UserRole          `json:"role"`
	ContactPreference ContactPreference `json:"contact_preference,omitempty"` // See DATA_SPINE_SPEC.md Section 2.3
}

// InvoiceExtraction represents the output of document analysis.
// See API_AND_TYPES_SPEC.md Section 3.1
type InvoiceExtraction struct {
	Vendor           string                  `json:"vendor"`
	Date             string                  `json:"date"`
	InvoiceNumber    string                  `json:"invoice_number"`
	TotalAmount      float64                 `json:"total_amount"`
	LineItems        []InvoiceExtractionItem `json:"line_items"`
	SuggestedWBSCode string                  `json:"suggested_wbs_code"`
	Confidence       float64                 `json:"confidence"`
}

// InvoiceExtractionItem represents a single line item in an invoice.
type InvoiceExtractionItem struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Total       float64 `json:"total"`
}

// GanttData represents the project schedule for the Gantt view.
// See API_AND_TYPES_SPEC.md Section 3.2
type GanttData struct {
	ProjectID        string      `json:"project_id"`
	CalculatedAt     string      `json:"calculated_at"`
	ProjectedEndDate string      `json:"projected_end_date"`
	CriticalPath     []string    `json:"critical_path"`
	Tasks            []GanttTask `json:"tasks"`
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

// AuthRequest is the payload for requesting a magic link.
type AuthRequest struct {
	Email string `json:"email"`
}

// AuthResponse is the generic response for auth requests.
type AuthResponse struct {
	Message string `json:"message"`
}

// User represents an internal user.
// See DATA_SPINE_SPEC.md Section 2.2
type User struct {
	ID        string   `json:"id"`
	OrgID     string   `json:"org_id"`
	Email     string   `json:"email"`
	Name      string   `json:"name"`
	Role      UserRole `json:"role"`
	CreatedAt string   `json:"created_at"`
}

// ChatMessage represents a single message in an agent session.
// See API_AND_TYPES_SPEC.md Section 4.4
type ChatMessage struct {
	ID        string      `json:"id"`         // UUID string
	ProjectID string      `json:"project_id"` // UUID string
	UserID    string      `json:"user_id"`    // UUID string
	Role      ChatRole    `json:"role"`
	Content   string      `json:"content"`
	ToolCalls interface{} `json:"tool_calls,omitempty"` // Use specific struct if available, else interface{}
	CreatedAt string      `json:"created_at"`           // ISO-8601 Timestamp
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
