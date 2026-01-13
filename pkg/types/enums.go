package types

// TaskStatus defines the lifecycle of a ProjectTask.
// See API_AND_TYPES_SPEC.md Section 1.1
type TaskStatus string

const (
	TaskStatusPending           TaskStatus = "Pending"
	TaskStatusReady             TaskStatus = "Ready"
	TaskStatusInProgress        TaskStatus = "In_Progress"
	TaskStatusInspectionPending TaskStatus = "Inspection_Pending"
	TaskStatusCompleted         TaskStatus = "Completed"
	TaskStatusBlocked           TaskStatus = "Blocked"
	TaskStatusDelayed           TaskStatus = "Delayed"
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

// ValidUserRole checks if a string is a valid UserRole enum value.
// See API_AND_TYPES_SPEC.md Section 1.2
func ValidUserRole(s string) bool {
	switch UserRole(s) {
	case UserRoleAdmin, UserRoleBuilder, UserRoleClient, UserRoleSubcontractor:
		return true
	default:
		return false
	}
}

// ArtifactType defines the visual components displayed in the Chat Orchestrator.
// See API_AND_TYPES_SPEC.md Section 1.3
type ArtifactType string

const (
	ArtifactTypeInvoice    ArtifactType = "Invoice"
	ArtifactTypeBudgetView ArtifactType = "Budget_View"
	ArtifactTypeGanttView  ArtifactType = "Gantt_View"
	ArtifactTypeDynamicUI  ArtifactType = "Dynamic_UI" // See API_AND_TYPES_SPEC.md Section 1.3
)

// ContactPreference defines the preferred communication channel for a contact.
// See DATA_SPINE_SPEC.md Section 2.3
type ContactPreference string

const (
	ContactPreferenceSMS   ContactPreference = "SMS"
	ContactPreferenceEmail ContactPreference = "Email"
	ContactPreferenceBoth  ContactPreference = "Both"
)

// DependencyType defines the relationship between two tasks.
// See DATA_SPINE_SPEC.md Section 3.4
type DependencyType string

const (
	DependencyTypeFS DependencyType = "FS"
	DependencyTypeSS DependencyType = "SS"
	DependencyTypeFF DependencyType = "FF"
	DependencyTypeSF DependencyType = "SF"
)

// InspectionResult defines the outcome of a QA check.
// See BACKEND_SCOPE.md Section 5.2
type InspectionResult string

const (
	InspectionResultPending     InspectionResult = "Pending"
	InspectionResultPassed      InspectionResult = "Passed"
	InspectionResultFailed      InspectionResult = "Failed"
	InspectionResultConditional InspectionResult = "Conditional"
)

// CommunicationDirection defines the flow of a message.
// See DATA_SPINE_SPEC.md Section 5.1
type CommunicationDirection string

const (
	CommunicationDirectionInbound  CommunicationDirection = "Inbound"
	CommunicationDirectionOutbound CommunicationDirection = "Outbound"
)

// CommunicationChannel defines the medium of a message.
type CommunicationChannel string

const (
	CommunicationChannelSMS   CommunicationChannel = "SMS"
	CommunicationChannelChat  CommunicationChannel = "Chat"
	CommunicationChannelEmail CommunicationChannel = "Email"
)

// NotificationType defines the category of a system alert.
// See DATA_SPINE_SPEC.md Section 5.2
type NotificationType string

const (
	NotificationTypeScheduleSlip  NotificationType = "Schedule_Slip"
	NotificationTypeInvoiceReady  NotificationType = "Invoice_Ready"
	NotificationTypeAssignmentNew NotificationType = "Assignment_New"
	NotificationTypeDailyBriefing NotificationType = "Daily_Briefing"
)

// NotificationStatus defines the read state of a notification.
type NotificationStatus string

const (
	NotificationStatusUnread    NotificationStatus = "Unread"
	NotificationStatusRead      NotificationStatus = "Read"
	NotificationStatusDismissed NotificationStatus = "Dismissed"
)

// ProcurementStatus defines the lifecycle status of a long-lead item.
// See BACKEND_SCOPE.md Section 4.2
type ProcurementStatus string

const (
	ProcurementStatusNotOrdered ProcurementStatus = "not_ordered"
	ProcurementStatusOrdered    ProcurementStatus = "ordered"
	ProcurementStatusInTransit  ProcurementStatus = "in_transit"
	ProcurementStatusDelivered  ProcurementStatus = "delivered"
	ProcurementStatusInstalled  ProcurementStatus = "installed"
)
