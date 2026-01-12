package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/colton/futurebuild/pkg/types"
)

// DirectoryService handles contact and assignment lookups.
// See API_AND_TYPES_SPEC.md Section 2.4
type DirectoryService struct {
	db *pgxpool.Pool
}

// NewDirectoryService creates a new DirectoryService.
func NewDirectoryService(db *pgxpool.Pool) *DirectoryService {
	return &DirectoryService{
		db: db,
	}
}

// GetContactForPhase resolves contact information for a project phase.
// It maps the phaseCode (e.g., "9.x") to a contact via the project_assignments table.
// Per DATA_SPINE_SPEC.md Section 3.5, wbs_phase_id is a VARCHAR containing the code.
// See PRODUCTION_PLAN.md Step 38
func (s *DirectoryService) GetContactForPhase(ctx context.Context, projectID, orgID uuid.UUID, phaseCode string) (*types.Contact, error) {
	// Logic: Resolve Phase Code to contact via simple JOIN
	query := `
		SELECT c.id, c.name, c.company, COALESCE(c.phone, ''), COALESCE(c.email, ''), c.role, c.contact_preference
		FROM contacts c
		JOIN project_assignments pa ON c.id = pa.contact_id
		WHERE pa.project_id = $1 
		  AND pa.wbs_phase_id = $2 
		  AND c.org_id = $3
	`

	var contact types.Contact
	var role string
	var preference string
	err := s.db.QueryRow(ctx, query, projectID, phaseCode, orgID).Scan(
		&contact.ID,
		&contact.Name,
		&contact.Company,
		&contact.Phone,
		&contact.Email,
		&role,
		&preference,
	)

	if err != nil {
		return nil, fmt.Errorf("contact for phase %s not found in project %s: %w", phaseCode, projectID, err)
	}

	if !types.ValidUserRole(role) {
		return nil, fmt.Errorf("invalid user role '%s' for contact %s", role, contact.ID)
	}

	contact.Role = types.UserRole(role)
	contact.ContactPreference = types.ContactPreference(preference)

	return &contact, nil
}
