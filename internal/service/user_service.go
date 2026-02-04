package service

import (
	"context"
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
func (s *UserService) ListOrgMembers(ctx context.Context, orgID uuid.UUID) ([]models.User, error) {
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
