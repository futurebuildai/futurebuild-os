package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/colton/futurebuild/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserService handles user-related business logic.
type UserService struct {
	db *pgxpool.Pool
}

// NewUserService creates a new user service.
func NewUserService(db *pgxpool.Pool) *UserService {
	return &UserService{db: db}
}

// ListOrgMembers returns all users belonging to the given organization.
// claimOrgID is the raw org_id from the JWT claims — it may be an internal UUID
// or a Clerk external_id (e.g. "org_xxx"). The method resolves either format.
func (s *UserService) ListOrgMembers(ctx context.Context, claimOrgID string) ([]models.User, error) {
	orgID, err := s.resolveOrgID(ctx, claimOrgID)
	if err != nil {
		return nil, fmt.Errorf("resolve org_id: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, org_id, email, name, role, created_at
		FROM users
		WHERE org_id = $1
		ORDER BY created_at DESC
	`, orgID)
	if err != nil {
		slog.Error("user_service: failed to list org members", "error", err, "org_id", orgID)
		return nil, err
	}
	defer rows.Close()

	members := make([]models.User, 0)
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.OrgID, &u.Email, &u.Name, &u.Role, &u.CreatedAt); err != nil {
			slog.Error("user_service: failed to scan user row", "error", err)
			return nil, err
		}
		members = append(members, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

// ResolveUserOrg looks up the user's org_id from their Clerk external_id.
// Used as a fallback when the JWT doesn't contain an org_id claim.
func (s *UserService) ResolveUserOrg(ctx context.Context, userExternalID string) (string, error) {
	var orgID uuid.UUID
	err := s.db.QueryRow(ctx,
		`SELECT org_id FROM users WHERE external_id = $1`,
		userExternalID,
	).Scan(&orgID)
	if err != nil {
		slog.Error("user_service: failed to resolve org from user external_id",
			"external_id", userExternalID, "error", err)
		return "", fmt.Errorf("user not found for external_id %q", userExternalID)
	}
	return orgID.String(), nil
}

// resolveOrgID converts a claim org_id to an internal UUID.
// Tries UUID parse first; falls back to external_id lookup on organizations table.
func (s *UserService) resolveOrgID(ctx context.Context, claimOrgID string) (uuid.UUID, error) {
	// Fast path: already a UUID
	if id, err := uuid.Parse(claimOrgID); err == nil {
		return id, nil
	}

	// Slow path: Clerk external_id lookup
	var id uuid.UUID
	err := s.db.QueryRow(ctx,
		`SELECT id FROM organizations WHERE external_id = $1`,
		claimOrgID,
	).Scan(&id)
	if err != nil {
		slog.Error("user_service: failed to resolve org external_id",
			"external_id", claimOrgID, "error", err)
		return uuid.Nil, fmt.Errorf("organization not found for external_id %q", claimOrgID)
	}

	return id, nil
}
