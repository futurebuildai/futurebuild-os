package service

import (
	"context"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/physics"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
)

// ProjectServicer defines the contract for project management operations.
type ProjectServicer interface {
	CreateProject(ctx context.Context, p *models.Project) error
	GetProject(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.Project, error)
	StreamActiveProjects(ctx context.Context, process ProjectProcessor) error
	ListProcurementItems(ctx context.Context, projectID, orgID uuid.UUID) ([]models.ProcurementItem, error)
}

// ScheduleServicer defines the contract for schedule and task management.
// Used by TaskHandler and Chat Orchestrator.
type ScheduleServicer interface {
	GetProjectSchedule(ctx context.Context, projectID, orgID uuid.UUID) (*ProjectScheduleSummary, error)
	GetAgentFocusTasks(ctx context.Context, projectID uuid.UUID) ([]models.ProjectTask, error)
	RecalculateSchedule(ctx context.Context, projectID, orgID uuid.UUID) (*physics.CPMResult, error)
	GetTask(ctx context.Context, taskID, projectID, orgID uuid.UUID) (*models.ProjectTask, error)
	UpdateTaskDuration(ctx context.Context, taskID, projectID, orgID uuid.UUID, overrideDays float64, reason string) error
	UpdateTaskStatus(ctx context.Context, taskID, projectID, orgID uuid.UUID, status types.TaskStatus) error
	CreateTaskProgress(ctx context.Context, projectID, taskID, userID uuid.UUID, percentComplete int, notes string) error
	CreateInspectionRecord(ctx context.Context, projectID, taskID uuid.UUID, inspectorName, result, notes string, inspectionDate time.Time) error
}

// InvoiceServicer defines the contract for invoice analysis and persistence.
type InvoiceServicer interface {
	AnalyzeInvoice(ctx context.Context, orgID uuid.UUID, docID uuid.UUID) (uuid.UUID, *types.InvoiceExtraction, error)
	SaveExtraction(ctx context.Context, projectID uuid.UUID, extraction *types.InvoiceExtraction, sourceDocID *uuid.UUID) (uuid.UUID, error)
}

// DocumentServicer defines the contract for document RAG (Chunk/Embed) operations.
type DocumentServicer interface {
	GetDocumentStatus(ctx context.Context, docID uuid.UUID) (string, int, error)
	IngestDocument(ctx context.Context, docID uuid.UUID) error
	ReprocessDocument(ctx context.Context, orgID, docID uuid.UUID) error
}

// DirectoryServicer defines contact and assignment lookups.
// Aligns with pkg/types.DirectoryService but exposed here for service-layer injection.
type DirectoryServicer interface {
	GetContactForPhase(ctx context.Context, projectID, orgID uuid.UUID, phaseCode string) (*types.Contact, error)
	GetProjectManager(ctx context.Context, projectID, orgID uuid.UUID) (*types.Contact, error)
}

// NotificationServicer defines outbound communication.
// Aligns with pkg/types.NotificationService.
type NotificationServicer interface {
	SendSMS(contactID string, message string) error
	SendEmail(to string, subject string, body string) error
}

// WeatherServicer defines integration for weather Data.
// Aligns with pkg/types.WeatherService.
type WeatherServicer interface {
	GetForecast(lat, long float64) (types.Forecast, error)
}

// VisionServicer defines the validation protocol service.
// Aligns with pkg/types.VisionService.
type VisionServicer interface {
	VerifyTask(ctx context.Context, imageURL string, taskDescription string) (bool, float64, error)
}

// GitHubServicer defines the contract for GitHub API operations.
// Used for Automated PR Review. See docs/AUTOMATED_PR_REVIEW_PRD.md
type GitHubServicer interface {
	FetchPRDiff(ctx context.Context, owner, repo string, prNumber int) (string, error)
	PostPRComment(ctx context.Context, owner, repo string, prNumber int, body string) error
}
