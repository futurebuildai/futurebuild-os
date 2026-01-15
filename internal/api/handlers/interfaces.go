package handlers

import (
	"context"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/physics"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
)

// ProjectServiceInterface defines the operations needed by ProjectHandler.
// This enables dependency injection for unit testing.
// See REMEDIATION_HANDLERS_TEST.md
type ProjectServiceInterface interface {
	CreateProject(ctx context.Context, p *models.Project) error
	GetProject(ctx context.Context, projectID, orgID uuid.UUID) (*models.Project, error)
}

// ScheduleServiceInterface defines the operations needed by TaskHandler.
// This enables dependency injection for unit testing.
// See REMEDIATION_HANDLERS_TEST.md
type ScheduleServiceInterface interface {
	GetTask(ctx context.Context, taskID, projectID, orgID uuid.UUID) (*models.ProjectTask, error)
	UpdateTaskDuration(ctx context.Context, taskID, projectID, orgID uuid.UUID, days float64, reason string) error
	UpdateTaskStatus(ctx context.Context, taskID, projectID, orgID uuid.UUID, status types.TaskStatus) error
	CreateTaskProgress(ctx context.Context, projectID, taskID, userID uuid.UUID, percent int, notes string) error
	CreateInspectionRecord(ctx context.Context, projectID, taskID uuid.UUID, inspector, result, notes string, date time.Time) error
	RecalculateSchedule(ctx context.Context, projectID, orgID uuid.UUID) (*physics.CPMResult, error)
}
