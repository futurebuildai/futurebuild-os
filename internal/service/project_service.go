package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/colton/futurebuild/internal/chaos"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	chaosInjector    chaos.Injector    // Optional: nil in production, used for self-healing tests
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

// NewProjectServiceWithChaos creates a ProjectService with chaos injection support.
// Used for self-healing integration tests (Tree Planting).
// SAFETY: Only use in test environments; production should use NewProjectService.
func NewProjectServiceWithChaos(db *pgxpool.Pool, injector chaos.Injector) *ProjectService {
	return &ProjectService{db: db, chaosInjector: injector}
}

// CreateProject persists a new project to the database.
// See DATA_SPINE_SPEC.md Section 3.1
// P1 Performance Fix: Enqueues hydration task after successful creation.
func (s *ProjectService) CreateProject(ctx context.Context, p *models.Project) error {
	// Chaos injection hook for self-healing tests
	if s.chaosInjector != nil {
		if shouldFail, err := s.chaosInjector.ShouldFail(ctx, "CreateProject"); shouldFail {
			return err
		}
	}

	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	if p.Status == "" {
		p.Status = models.ProjectStatusPreconstruction
	}

	query := `
		INSERT INTO projects (id, org_id, name, address, permit_issued_date, target_end_date, gsf, status,
			bedrooms, bathrooms, stories, lot_size, foundation_type, topography, soil_conditions)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := s.db.Exec(ctx, query,
		p.ID, p.OrgID, p.Name, p.Address, p.PermitIssuedDate, p.TargetEndDate, p.GSF, p.Status,
		p.Bedrooms, p.Bathrooms, p.Stories, p.LotSize, p.FoundationType, p.Topography, p.SoilConditions)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// L7 Fix: Return typed sentinel
			return fmt.Errorf("%w: project with name '%s' already exists", types.ErrConflict, p.Name)
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
// See DATA_SPINE_SPEC.md Section 3.1
func (s *ProjectService) GetProject(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.Project, error) {
	query := `
		SELECT id, org_id, name, address, permit_issued_date, target_end_date, gsf, status,
			bedrooms, bathrooms, stories, lot_size, foundation_type, topography, soil_conditions,
			completed_at, completed_by
		FROM projects
		WHERE id = $1 AND org_id = $2
	`
	var p models.Project
	err := s.db.QueryRow(ctx, query, id, orgID).Scan(
		&p.ID, &p.OrgID, &p.Name, &p.Address, &p.PermitIssuedDate, &p.TargetEndDate, &p.GSF, &p.Status,
		&p.Bedrooms, &p.Bathrooms, &p.Stories, &p.LotSize, &p.FoundationType, &p.Topography, &p.SoilConditions,
		&p.CompletedAt, &p.CompletedBy)
	if err != nil {
		// L7 Fix: Return typed sentinel
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: project %s", types.ErrNotFound, id)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &p, nil
}

// ProjectProcessor is a callback function for processing projects one-by-one.
// Uses streaming to prevent OOM at scale.
type ProjectProcessor func(p models.Project) error

// ListActiveProjects fetches all projects in Active or Preconstruction status.
// Used by DailyFocusAgent for batch processing.
// See PRODUCTION_PLAN.md Step 49 (Service Layer Pattern)
// DEPRECATED: Use StreamActiveProjects for O(1) memory at scale.
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

// StreamActiveProjects iterates through active projects via callback.
// O(1) memory - only one project loaded at a time.
// P1 Scalability Fix: Prevents OOM at 100K+ projects.
// See PRODUCTION_PLAN.md: DailyFocusAgent O(N) Memory Elimination
func (s *ProjectService) StreamActiveProjects(ctx context.Context, process ProjectProcessor) error {
	query := `
		SELECT id, org_id, name, address, permit_issued_date, target_end_date, status
		FROM projects
		WHERE status IN ('Active', 'Preconstruction')
		ORDER BY id
	`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to stream active projects: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.OrgID, &p.Name, &p.Address, &p.PermitIssuedDate, &p.TargetEndDate, &p.Status); err != nil {
			return fmt.Errorf("failed to scan project: %w", err)
		}
		if err := process(p); err != nil {
			return fmt.Errorf("failed to process project %s: %w", p.ID, err)
		}
	}
	return rows.Err()
}

// ListProcurementItems fetches all procurement items for a project.
// See BACKEND_SCOPE.md Section 2.5 (Long-Lead Procurement Items)
// Multi-tenancy enforced via JOIN on projects.org_id (single query, no N+1).
func (s *ProjectService) ListProcurementItems(ctx context.Context, projectID, orgID uuid.UUID) ([]models.ProcurementItem, error) {
	// Single query with multi-tenancy enforcement via JOIN
	// Eliminates the double round-trip of calling GetProject first (P1-3 fix)
	query := `
		SELECT 
			pi.id, pt.project_id, pt.wbs_code, pi.name, 
			pi.lead_time_weeks, pi.status, pi.calculated_order_date, 
			pi.expected_delivery_date, pi.created_at
		FROM procurement_items pi
		JOIN project_tasks pt ON pi.project_task_id = pt.id
		JOIN projects p ON pt.project_id = p.id
		WHERE pt.project_id = $1 AND p.org_id = $2
		ORDER BY pi.created_at DESC
	`
	rows, err := s.db.Query(ctx, query, projectID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list procurement items: %w", err)
	}
	defer rows.Close()

	// P1-4 fix: Initialize as empty slice, not nil, to return [] instead of null in JSON
	items := make([]models.ProcurementItem, 0)
	for rows.Next() {
		var item models.ProcurementItem
		err := rows.Scan(
			&item.ID, &item.ProjectID, &item.WBSCode, &item.ItemName,
			&item.LeadTimeWeeks, &item.Status, &item.CalculatedOrderDate,
			&item.ExpectedDeliveryDate, &item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan procurement item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
