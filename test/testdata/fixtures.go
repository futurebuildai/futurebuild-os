// Package testdata provides factory functions for test data creation.
// See PRODUCTION_PLAN.md Technical Debt Remediation (P2) Section A.
//
// ENGINEERING STANDARD: Use these factories instead of raw SQL INSERT statements.
// This ensures tests remain valid as the database schema evolves.
package testdata

import (
	"context"
	"fmt"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Functional Options Pattern ---

// ProjectOption configures optional Project fields.
type ProjectOption func(*models.Project)

// WithProjectStatus sets the project status.
func WithProjectStatus(status string) ProjectOption {
	return func(p *models.Project) {
		p.Status = models.ProjectStatus(status)
	}
}

// WithPermitDate sets the project permit issued date.
func WithPermitDate(t time.Time) ProjectOption {
	return func(p *models.Project) {
		p.PermitIssuedDate = &t
	}
}

// WithAddress sets the project address.
func WithAddress(addr string) ProjectOption {
	return func(p *models.Project) {
		p.Address = addr
	}
}

// TaskOption configures optional ProjectTask fields.
type TaskOption func(*models.ProjectTask)

// WithEarlyStart sets the task's early start date.
func WithEarlyStart(t time.Time) TaskOption {
	return func(task *models.ProjectTask) {
		task.EarlyStart = &t
	}
}

// WithCriticalPath sets whether the task is on the critical path.
func WithCriticalPath(onCriticalPath bool) TaskOption {
	return func(task *models.ProjectTask) {
		task.IsOnCriticalPath = onCriticalPath
	}
}

// WithTaskStatus sets the task status.
func WithTaskStatus(status types.TaskStatus) TaskOption {
	return func(task *models.ProjectTask) {
		task.Status = status
	}
}

// --- Factory Functions ---

// NewTestOrganization creates a test organization and returns its ID.
// The organization is persisted to the database immediately.
func NewTestOrganization(ctx context.Context, db *pgxpool.Pool, name string) (uuid.UUID, error) {
	id := uuid.New()
	_, err := db.Exec(ctx, `INSERT INTO organizations (id, name) VALUES ($1, $2)`, id, name)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create test organization: %w", err)
	}
	return id, nil
}

// NewTestProject creates a test project with sensible defaults.
// Optional parameters can be set via ProjectOption functions.
func NewTestProject(ctx context.Context, db *pgxpool.Pool, orgID uuid.UUID, name string, opts ...ProjectOption) (*models.Project, error) {
	project := &models.Project{
		ID:     uuid.New(),
		OrgID:  orgID,
		Name:   name,
		Status: models.ProjectStatusActive,
	}

	// Apply optional configurations
	for _, opt := range opts {
		opt(project)
	}

	query := `
		INSERT INTO projects (id, org_id, name, status, permit_issued_date, address)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := db.Exec(ctx, query, project.ID, project.OrgID, project.Name, project.Status, project.PermitIssuedDate, project.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to create test project: %w", err)
	}
	return project, nil
}

// NewTestTask creates a test task with sensible defaults.
// Optional parameters can be set via TaskOption functions.
func NewTestTask(ctx context.Context, db *pgxpool.Pool, projectID uuid.UUID, wbsCode, name string, opts ...TaskOption) (*models.ProjectTask, error) {
	task := &models.ProjectTask{
		ID:        uuid.New(),
		ProjectID: projectID,
		WBSCode:   wbsCode,
		Name:      name,
		Status:    types.TaskStatusPending,
	}

	// Apply optional configurations
	for _, opt := range opts {
		opt(task)
	}

	query := `
		INSERT INTO project_tasks (id, project_id, wbs_code, name, status, early_start, is_on_critical_path)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := db.Exec(ctx, query, task.ID, task.ProjectID, task.WBSCode, task.Name, task.Status, task.EarlyStart, task.IsOnCriticalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create test task: %w", err)
	}
	return task, nil
}

// NewTestProcurementItem creates a test procurement item linked to a task.
func NewTestProcurementItem(ctx context.Context, db *pgxpool.Pool, taskID uuid.UUID, name string, leadTimeWeeks int) (uuid.UUID, error) {
	id := uuid.New()
	query := `
		INSERT INTO procurement_items (id, project_task_id, name, lead_time_weeks, status)
		VALUES ($1, $2, $3, $4, 'pending')
	`
	_, err := db.Exec(ctx, query, id, taskID, name, leadTimeWeeks)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create test procurement item: %w", err)
	}
	return id, nil
}

// NewTestContact creates a test contact with the given details.
func NewTestContact(ctx context.Context, db *pgxpool.Pool, name, phone, email, role, preference string) (uuid.UUID, error) {
	id := uuid.New()
	query := `
		INSERT INTO contacts (id, name, phone, email, role, contact_preference)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := db.Exec(ctx, query, id, name, phone, email, role, preference)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create test contact: %w", err)
	}
	return id, nil
}

// NewTestProjectAssignment assigns a contact to a project phase.
// Returns nil error if wbs_phases table is not seeded (non-fatal for tests).
func NewTestProjectAssignment(ctx context.Context, db *pgxpool.Pool, projectID, contactID uuid.UUID, phaseCode string) error {
	query := `
		INSERT INTO project_assignments (project_id, contact_id, assigned_phase_id)
		SELECT $1, $2, wp.id FROM wbs_phases wp WHERE wp.code = $3 LIMIT 1
	`
	_, err := db.Exec(ctx, query, projectID, contactID, phaseCode)
	// Assignment may fail if wbs_phases not seeded - this is tolerable in some tests
	return err
}

// CleanupTestProject removes all test data associated with a project.
// Clean up in reverse order of creation to respect foreign key constraints.
func CleanupTestProject(ctx context.Context, db *pgxpool.Pool, projectID uuid.UUID, taskIDs ...uuid.UUID) {
	// Clean up communication logs
	_, _ = db.Exec(ctx, `DELETE FROM communication_logs WHERE project_id = $1`, projectID)

	// Clean up procurement items for each task
	for _, taskID := range taskIDs {
		_, _ = db.Exec(ctx, `DELETE FROM procurement_items WHERE project_task_id = $1`, taskID)
	}

	// Clean up tasks
	_, _ = db.Exec(ctx, `DELETE FROM project_tasks WHERE project_id = $1`, projectID)

	// Clean up assignments
	_, _ = db.Exec(ctx, `DELETE FROM project_assignments WHERE project_id = $1`, projectID)

	// Clean up project
	_, _ = db.Exec(ctx, `DELETE FROM projects WHERE id = $1`, projectID)

	// Note: Organizations and contacts are intentionally left behind
	// to avoid FK issues and to support shared test fixtures
}
