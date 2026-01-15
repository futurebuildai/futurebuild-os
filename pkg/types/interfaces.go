package types

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// --- Handler Service Interfaces ---
// These interfaces enable dependency injection for testable handlers.
// See REMEDIATION_HANDLERS_TEST.md

// ProjectService defines the operations for project management.
type ProjectService interface {
	CreateProject(ctx context.Context, p interface{}) error
	GetProject(ctx context.Context, projectID, orgID uuid.UUID) (interface{}, error)
}

// ScheduleService defines the operations for task/schedule management.
// Used by TaskHandler for all CPM-related operations.
type ScheduleService interface {
	GetTask(ctx context.Context, taskID, projectID, orgID uuid.UUID) (interface{}, error)
	UpdateTaskDuration(ctx context.Context, taskID, projectID, orgID uuid.UUID, days float64, reason string) error
	UpdateTaskStatus(ctx context.Context, taskID, projectID, orgID uuid.UUID, status TaskStatus) error
	CreateTaskProgress(ctx context.Context, projectID, taskID, userID uuid.UUID, percent int, notes string) error
	CreateInspectionRecord(ctx context.Context, projectID, taskID uuid.UUID, inspector, result, notes string, date time.Time) error
	RecalculateSchedule(ctx context.Context, projectID, orgID uuid.UUID) (interface{}, error)
}

// --- External Service Interfaces ---

// WeatherService defines the integration for the SWIM Model.
// See API_AND_TYPES_SPEC.md Section 2.1
type WeatherService interface {
	GetForecast(lat, long float64) (Forecast, error)
}

// VisionService defines the Validation Protocol service.
// See API_AND_TYPES_SPEC.md Section 2.2
type VisionService interface {
	// VerifyTask returns (is_verified, confidence_score, error)
	// Context is required for AI inference timeout/cancellation control.
	VerifyTask(ctx context.Context, imageURL string, taskDescription string) (bool, float64, error)
}

// NotificationService defines the outbound communication service.
// See API_AND_TYPES_SPEC.md Section 2.3
type NotificationService interface {
	SendSMS(contactID string, message string) error
	SendEmail(to string, subject string, body string) error
}

// DirectoryService defines contact and assignment lookups.
// See API_AND_TYPES_SPEC.md Section 2.4
type DirectoryService interface {
	GetContactForPhase(ctx context.Context, projectID, orgID uuid.UUID, phaseCode string) (*Contact, error)
}
