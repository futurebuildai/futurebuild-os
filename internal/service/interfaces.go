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
// Used by TaskHandler, ScheduleHandler, and Chat Orchestrator.
type ScheduleServicer interface {
	GetProjectSchedule(ctx context.Context, projectID, orgID uuid.UUID) (*ProjectScheduleSummary, error)
	GetGanttData(ctx context.Context, projectID, orgID uuid.UUID) (*types.GanttData, error) // Phase 14: Full Gantt view data
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
	GetInvoice(ctx context.Context, invoiceID uuid.UUID, orgID uuid.UUID) (*models.Invoice, error)
	UpdateInvoiceItems(ctx context.Context, invoiceID uuid.UUID, orgID uuid.UUID, items []models.LineItem) (*models.Invoice, error)
	ApproveInvoice(ctx context.Context, invoiceID uuid.UUID, orgID uuid.UUID, approverID string) (*models.Invoice, error)
	RejectInvoice(ctx context.Context, invoiceID uuid.UUID, orgID uuid.UUID, rejectorID string, reason string) (*models.Invoice, error)
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

// AssetServicer defines the contract for project asset (photo) operations.
// See STEP_84_FIELD_FEEDBACK.md Section 2, STEP_85_VISION_BADGES.md Section 2
type AssetServicer interface {
	GetAssetStatus(ctx context.Context, assetID uuid.UUID, orgID uuid.UUID) (*models.ProjectAsset, error)
	ListProjectAssets(ctx context.Context, projectID uuid.UUID, orgID uuid.UUID) ([]models.ProjectAsset, error)
}

// ConfigServicer defines the contract for org-level configuration operations.
// See STEP_87_CONFIG_PERSISTENCE.md Section 2
type ConfigServicer interface {
	GetConfig(ctx context.Context, orgID uuid.UUID) (*models.BusinessConfig, error)
	UpdateConfig(ctx context.Context, orgID uuid.UUID, speedMultiplier float64, workDays []int) (*models.BusinessConfig, error)
}

// ThreadServicer defines the contract for conversation thread operations.
type ThreadServicer interface {
	CreateThread(ctx context.Context, projectID, orgID, userID uuid.UUID, title string) (*models.Thread, error)
	CreateGeneralThread(ctx context.Context, projectID uuid.UUID) (*models.Thread, error)
	ListThreads(ctx context.Context, projectID, orgID uuid.UUID, includeArchived bool) ([]models.Thread, error)
	GetThread(ctx context.Context, threadID, projectID, orgID uuid.UUID) (*models.Thread, error)
	ArchiveThread(ctx context.Context, threadID, projectID, orgID uuid.UUID) error
	UnarchiveThread(ctx context.Context, threadID, projectID, orgID uuid.UUID) error
	GetOrCreateGeneralThread(ctx context.Context, projectID uuid.UUID) (*models.Thread, error)
	GetThreadMessages(ctx context.Context, threadID, projectID, orgID uuid.UUID, limit int) ([]models.ChatMessage, error)
}

// CompletionServicer defines the contract for project completion lifecycle operations.
type CompletionServicer interface {
	CompleteProject(ctx context.Context, projectID, orgID, userID uuid.UUID, notes string) (*models.CompletionReport, error)
	GetCompletionReport(ctx context.Context, projectID, orgID uuid.UUID) (*models.CompletionReport, error)
}

// UserServicer defines the contract for user management operations.
type UserServicer interface {
	ListOrgMembers(ctx context.Context, claimOrgID string) ([]models.User, error)
	ResolveUserOrg(ctx context.Context, userExternalID string) (string, error)
}

// GitHubServicer defines the contract for GitHub API operations.
// Used for Automated PR Review. See docs/AUTOMATED_PR_REVIEW_PRD.md
type GitHubServicer interface {
	FetchPRDiff(ctx context.Context, owner, repo string, prNumber int) (string, error)
	PostPRComment(ctx context.Context, owner, repo string, prNumber int, body string) error
}

// FeedServicer defines the contract for portfolio feed operations.
// See FRONTEND_V2_SPEC.md §5.1, §5.5
// Note: Primary definition is in feed_service.go with the implementation.
// Listed here for discoverability alongside other service interfaces.
