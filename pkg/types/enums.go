package types

// TaskStatus defines the lifecycle of a ProjectTask.
// See API_AND_TYPES_SPEC.md Section 1.1
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "Pending"
	TaskStatusReady      TaskStatus = "Ready"
	TaskStatusInProgress TaskStatus = "In_Progress"
	TaskStatusCompleted  TaskStatus = "Completed"
	TaskStatusBlocked    TaskStatus = "Blocked"
	TaskStatusDelayed    TaskStatus = "Delayed"
)

// UserRole defines the permissions and identity of a PortalUser or User.
// See API_AND_TYPES_SPEC.md Section 1.2
type UserRole string

const (
	UserRoleAdmin         UserRole = "Admin"
	UserRoleBuilder       UserRole = "Builder"
	UserRoleClient        UserRole = "Client"
	UserRoleSubcontractor UserRole = "Subcontractor"
)

// ArtifactType defines the visual components displayed in the Chat Orchestrator.
// See API_AND_TYPES_SPEC.md Section 1.3
type ArtifactType string

const (
	ArtifactTypeInvoice    ArtifactType = "Invoice"
	ArtifactTypeBudgetView ArtifactType = "Budget_View"
	ArtifactTypeGanttView  ArtifactType = "Gantt_View"
)
