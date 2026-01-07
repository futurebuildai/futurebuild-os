package service

import (
	"context"
	"fmt"

	"github.com/colton/futurebuild/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectService struct {
	db *pgxpool.Pool
}

func NewProjectService(db *pgxpool.Pool) *ProjectService {
	return &ProjectService{db: db}
}

// CreateProject persists a new project to the database.
// // See DATA_SPINE_SPEC.md Section 3.1
func (s *ProjectService) CreateProject(ctx context.Context, p *models.Project) error {
	// Antagonistic Check: Prevent duplicate names in same Org
	var exists bool
	dupErr := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM projects WHERE org_id = $1 AND name = $2)", p.OrgID, p.Name).Scan(&exists)
	if dupErr == nil && exists {
		return fmt.Errorf("project with name '%s' already exists in this organization", p.Name)
	}

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
		return fmt.Errorf("failed to create project: %w", err)
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
