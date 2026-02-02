package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InviteService handles user invitation operations for invite-only onboarding.
// See LAUNCH_STRATEGY.md Task B2: User Invite Flow.
type InviteService struct {
	db *pgxpool.Pool
}

// NewInviteService creates a new invite service.
func NewInviteService(db *pgxpool.Pool) *InviteService {
	return &InviteService{db: db}
}

// Invitation represents a pending user invitation.
type Invitation struct {
	ID         uuid.UUID       `json:"id"`
	OrgID      uuid.UUID       `json:"org_id"`
	Email      string          `json:"email"`
	Role       types.UserRole  `json:"role"`
	ExpiresAt  time.Time       `json:"expires_at"`
	AcceptedAt *time.Time      `json:"accepted_at,omitempty"`
	CreatedBy  uuid.UUID       `json:"created_by"`
	CreatedAt  time.Time       `json:"created_at"`
}

// CreateInvitationInput contains the data needed to create an invitation.
type CreateInvitationInput struct {
	OrgID     uuid.UUID
	Email     string
	Role      types.UserRole
	CreatedBy uuid.UUID
}

// CreateInvitation creates a new invitation and returns the raw token for email delivery.
// The token is returned only once; only the hash is stored in the database.
func (s *InviteService) CreateInvitation(ctx context.Context, input CreateInvitationInput) (rawToken string, inv *Invitation, err error) {
	// Validate role
	if input.Role == "" {
		input.Role = types.UserRoleBuilder
	}
	if !types.ValidUserRole(string(input.Role)) {
		return "", nil, fmt.Errorf("invalid role: %s", input.Role)
	}

	// Check if email already has a pending invitation
	var existingCount int
	checkQuery := `
		SELECT COUNT(*) FROM invitations
		WHERE org_id = $1 AND email = $2 AND accepted_at IS NULL AND expires_at > NOW()
	`
	if err := s.db.QueryRow(ctx, checkQuery, input.OrgID, input.Email).Scan(&existingCount); err != nil {
		return "", nil, fmt.Errorf("failed to check existing invitations: %w", err)
	}
	if existingCount > 0 {
		return "", nil, fmt.Errorf("pending invitation already exists for this email")
	}

	// Check if user already exists
	var existingUserCount int
	userCheckQuery := `SELECT COUNT(*) FROM users WHERE email = $1`
	if err := s.db.QueryRow(ctx, userCheckQuery, input.Email).Scan(&existingUserCount); err != nil {
		return "", nil, fmt.Errorf("failed to check existing users: %w", err)
	}
	if existingUserCount > 0 {
		return "", nil, fmt.Errorf("user with this email already exists")
	}

	// Generate token
	rawToken, err = generateToken()
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}
	tokenHash := hashToken(rawToken)

	// Create invitation with 24-hour expiry
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	invID := uuid.New()

	insertQuery := `
		INSERT INTO invitations (id, org_id, email, role, token_hash, expires_at, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING created_at
	`
	var createdAt time.Time
	err = s.db.QueryRow(ctx, insertQuery,
		invID, input.OrgID, input.Email, string(input.Role), tokenHash, expiresAt, input.CreatedBy,
	).Scan(&createdAt)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	inv = &Invitation{
		ID:        invID,
		OrgID:     input.OrgID,
		Email:     input.Email,
		Role:      input.Role,
		ExpiresAt: expiresAt,
		CreatedBy: input.CreatedBy,
		CreatedAt: createdAt,
	}

	return rawToken, inv, nil
}

// ListInvitations returns all pending invitations for an organization.
func (s *InviteService) ListInvitations(ctx context.Context, orgID uuid.UUID) ([]Invitation, error) {
	query := `
		SELECT id, org_id, email, role, expires_at, accepted_at, created_by, created_at
		FROM invitations
		WHERE org_id = $1 AND accepted_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := s.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list invitations: %w", err)
	}
	defer rows.Close()

	var invitations []Invitation
	for rows.Next() {
		var inv Invitation
		var roleStr string
		if err := rows.Scan(
			&inv.ID, &inv.OrgID, &inv.Email, &roleStr, &inv.ExpiresAt, &inv.AcceptedAt, &inv.CreatedBy, &inv.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan invitation: %w", err)
		}
		inv.Role = types.UserRole(roleStr)
		invitations = append(invitations, inv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating invitations: %w", err)
	}

	return invitations, nil
}

// RevokeInvitation deletes a pending invitation.
func (s *InviteService) RevokeInvitation(ctx context.Context, inviteID, orgID uuid.UUID) error {
	deleteQuery := `
		DELETE FROM invitations
		WHERE id = $1 AND org_id = $2 AND accepted_at IS NULL
	`
	result, err := s.db.Exec(ctx, deleteQuery, inviteID, orgID)
	if err != nil {
		return fmt.Errorf("failed to revoke invitation: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("invitation not found or already accepted")
	}
	return nil
}

// AcceptInvitation validates an invitation token, creates the user, and returns the new user.
// This method is transactional.
func (s *InviteService) AcceptInvitation(ctx context.Context, rawToken, name string) (*models.User, error) {
	tokenHash := hashToken(rawToken)

	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Find and validate invitation
	var inv Invitation
	var roleStr string
	findQuery := `
		SELECT id, org_id, email, role, expires_at, accepted_at
		FROM invitations
		WHERE token_hash = $1
	`
	err = tx.QueryRow(ctx, findQuery, tokenHash).Scan(
		&inv.ID, &inv.OrgID, &inv.Email, &roleStr, &inv.ExpiresAt, &inv.AcceptedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid invitation token")
	}
	inv.Role = types.UserRole(roleStr)

	// Validate invitation status
	if inv.AcceptedAt != nil {
		return nil, fmt.Errorf("invitation already used")
	}
	if time.Now().After(inv.ExpiresAt) {
		return nil, fmt.Errorf("invitation expired")
	}

	// Create user
	userID := uuid.New()
	createUserQuery := `
		INSERT INTO users (id, org_id, email, name, role, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING created_at
	`
	var createdAt time.Time
	err = tx.QueryRow(ctx, createUserQuery,
		userID, inv.OrgID, inv.Email, name, string(inv.Role),
	).Scan(&createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Mark invitation as accepted
	acceptQuery := `UPDATE invitations SET accepted_at = NOW() WHERE id = $1`
	_, err = tx.Exec(ctx, acceptQuery, inv.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to mark invitation accepted: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &models.User{
		ID:        userID,
		OrgID:     inv.OrgID,
		Email:     inv.Email,
		Name:      name,
		Role:      models.UserRole(inv.Role),
		CreatedAt: createdAt,
	}, nil
}

// GetInvitationByToken retrieves invitation details by token (for profile setup).
func (s *InviteService) GetInvitationByToken(ctx context.Context, rawToken string) (*Invitation, error) {
	tokenHash := hashToken(rawToken)

	var inv Invitation
	var roleStr string
	query := `
		SELECT id, org_id, email, role, expires_at, accepted_at, created_by, created_at
		FROM invitations
		WHERE token_hash = $1
	`
	err := s.db.QueryRow(ctx, query, tokenHash).Scan(
		&inv.ID, &inv.OrgID, &inv.Email, &roleStr, &inv.ExpiresAt, &inv.AcceptedAt, &inv.CreatedBy, &inv.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invitation not found")
	}
	inv.Role = types.UserRole(roleStr)

	return &inv, nil
}

// generateToken creates a cryptographically secure 32-byte random string.
func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hashToken returns the SHA-256 hash of a plaintext token.
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
