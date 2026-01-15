package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/colton/futurebuild/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HydrationEnqueuer enqueues project hydration tasks.
// Abstracts the async task queue for testability and to avoid import cycles.
// P1 Performance Fix: Enables event-driven hydration.
type HydrationEnqueuer interface {
	EnqueueHydration(ctx context.Context, projectID uuid.UUID) error
}

type ProjectService struct {
	db               *pgxpool.Pool
	hydrationEnqueue HydrationEnqueuer // Optional: nil means no async hydration
}

// NewProjectService creates a new ProjectService instance.
func NewProjectService(db *pgxpool.Pool) *ProjectService {
	return &ProjectService{db: db}
}

// NewProjectServiceWithHydration creates a ProjectService with async hydration support.
// P1 Performance Fix: Enables event-driven hydration on project creation.
func NewProjectServiceWithHydration(db *pgxpool.Pool, enqueue HydrationEnqueuer) *ProjectService {
	return &ProjectService{db: db, hydrationEnqueue: enqueue}
}

// CreateProject persists a new project to the database.
// See DATA_SPINE_SPEC.md Section 3.1
// P1 Performance Fix: Enqueues hydration task after successful creation.
func (s *ProjectService) CreateProject(ctx context.Context, p *models.Project) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	if p.Status == "" {
		p.Status = models.ProjectStatusPreconstruction
	}

	query := `
		INSERT INTO projects (id, org_id, name, address, permit_issued_date, target_end_date, gsf, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := s.db.Exec(ctx, query,
		p.ID, p.OrgID, p.Name, p.Address, p.PermitIssuedDate, p.TargetEndDate, p.GSF, p.Status)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("project with name '%s' already exists in this organization", p.Name)
		}
		return fmt.Errorf("failed to create project: %w", err)
	}

	// Event-driven hydration: Enqueue task to populate procurement_items
	// P1 Performance Fix: Replaces inefficient cron-swept hydration
	if s.hydrationEnqueue != nil {
		if err := s.hydrationEnqueue.EnqueueHydration(ctx, p.ID); err != nil {
			// Log but don't fail - hydration can be retried manually
			slog.Error("failed to enqueue hydration task", "project_id", p.ID, "error", err)
		} else {
			slog.Info("Enqueued hydration task for new project", "project_id", p.ID)
		}
	}

	return nil
}

// GetProject retrieves a project by ID, ensuring multi-tenancy via orgID.
// // See DATA_SPINE_SPEC.md Section 3.1
func (s *ProjectService) GetProject(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.Project, error) {
	query := `
		SELECT id, org_id, name, address, permit_issued_date, target_end_date, gsf, status
		FROM projects
		WHERE id = $1 AND org_id = $2
	`
	var p models.Project
	err := s.db.QueryRow(ctx, query, id, orgID).Scan(
		&p.ID, &p.OrgID, &p.Name, &p.Address, &p.PermitIssuedDate, &p.TargetEndDate, &p.GSF, &p.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &p, nil
}

// ListActiveProjects fetches all projects in Active or Preconstruction status.
// Used by DailyFocusAgent for batch processing.
// See PRODUCTION_PLAN.md Step 49 (Service Layer Pattern)
func (s *ProjectService) ListActiveProjects(ctx context.Context) ([]models.Project, error) {
	query := `
		SELECT id, org_id, name, address, permit_issued_date, target_end_date, status
		FROM projects
		WHERE status IN ('Active', 'Preconstruction')
	`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list active projects: %w", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		err := rows.Scan(&p.ID, &p.OrgID, &p.Name, &p.Address, &p.PermitIssuedDate, &p.TargetEndDate, &p.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}
