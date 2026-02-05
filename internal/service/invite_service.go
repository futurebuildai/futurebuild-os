package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InviteService handles user invitation operations for invite-only onboarding.
// See LAUNCH_STRATEGY.md Task B2: User Invite Flow.
type InviteService struct {
	db    *pgxpool.Pool
	clerk *ClerkClient
}

// NewInviteService creates a new invite service.
// clerk may be nil if Clerk Backend API is not configured.
func NewInviteService(db *pgxpool.Pool, clerk *ClerkClient) *InviteService {
	return &InviteService{db: db, clerk: clerk}
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

// RevokeInvitation deletes an invitation regardless of acceptance state.
// An admin can revoke any invitation in their org.
func (s *InviteService) RevokeInvitation(ctx context.Context, inviteID, orgID uuid.UUID) error {
	deleteQuery := `
		DELETE FROM invitations
		WHERE id = $1 AND org_id = $2
	`
	result, err := s.db.Exec(ctx, deleteQuery, inviteID, orgID)
	if err != nil {
		return fmt.Errorf("failed to revoke invitation: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("invitation not found")
	}
	return nil
}

// AcceptInvitation validates an invitation token, creates a Clerk user account,
// adds the user to the Clerk organization, upserts the local DB user, and marks
// the invite as accepted (last, so failures leave the invite retryable).
func (s *InviteService) AcceptInvitation(ctx context.Context, rawToken, name, password string) (*models.User, error) {
	tokenHash := hashToken(rawToken)

	slog.Info("invite: [1/6] looking up invitation by token hash")

	// Find and validate invitation
	var inv Invitation
	var roleStr string
	findQuery := `
		SELECT id, org_id, email, role, expires_at, accepted_at
		FROM invitations
		WHERE token_hash = $1
	`
	err := s.db.QueryRow(ctx, findQuery, tokenHash).Scan(
		&inv.ID, &inv.OrgID, &inv.Email, &roleStr, &inv.ExpiresAt, &inv.AcceptedAt,
	)
	if err != nil {
		slog.Error("invite: [1/6] token lookup FAILED", "error", err)
		return nil, fmt.Errorf("invalid invitation token")
	}
	inv.Role = types.UserRole(roleStr)

	slog.Info("invite: [1/6] invitation found",
		"invite_id", inv.ID, "email", inv.Email, "org_id", inv.OrgID,
		"role", roleStr, "expires_at", inv.ExpiresAt, "already_accepted", inv.AcceptedAt != nil)

	// Validate invitation status
	if inv.AcceptedAt != nil {
		return nil, fmt.Errorf("invitation already used")
	}
	if time.Now().After(inv.ExpiresAt) {
		slog.Warn("invite: invitation expired", "expires_at", inv.ExpiresAt, "now", time.Now())
		return nil, fmt.Errorf("invitation expired")
	}

	// Look up the Clerk org ID (external_id) for this organization
	slog.Info("invite: [2/6] looking up org external_id", "org_id", inv.OrgID)
	var clerkOrgID *string
	orgQuery := `SELECT external_id FROM organizations WHERE id = $1`
	err = s.db.QueryRow(ctx, orgQuery, inv.OrgID).Scan(&clerkOrgID)
	if err != nil {
		slog.Error("invite: [2/6] org lookup FAILED", "org_id", inv.OrgID, "error", err)
		return nil, fmt.Errorf("failed to look up organization: %w", err)
	}

	clerkOrgStr := "<nil>"
	if clerkOrgID != nil {
		clerkOrgStr = *clerkOrgID
	}
	slog.Info("invite: [2/6] org external_id resolved", "org_id", inv.OrgID, "clerk_org_id", clerkOrgStr)

	// Split name into first/last for Clerk
	firstName, lastName := splitName(name)

	var clerkUserID string

	// Create Clerk user and add to org (if Clerk client is configured)
	if s.clerk != nil {
		slog.Info("invite: [3/6] creating Clerk user", "email", inv.Email, "first_name", firstName, "last_name", lastName)

		clerkUser, err := s.clerk.CreateUser(ctx, inv.Email, password, firstName, lastName)
		if err != nil {
			slog.Error("invite: [3/6] Clerk CreateUser FAILED", "email", inv.Email, "error", err)
			return nil, fmt.Errorf("failed to create account: %w", err)
		}
		clerkUserID = clerkUser.ID
		slog.Info("invite: [3/6] Clerk user created OK", "clerk_user_id", clerkUserID, "email", inv.Email)

		// Add to Clerk organization if the org has a Clerk external_id
		if clerkOrgID != nil && *clerkOrgID != "" {
			clerkRole := MapInternalRoleToClerk(string(inv.Role))
			slog.Info("invite: [4/6] adding Clerk org membership",
				"clerk_user_id", clerkUserID, "clerk_org_id", *clerkOrgID, "clerk_role", clerkRole)

			if err := s.clerk.AddOrgMembership(ctx, *clerkOrgID, clerkUserID, clerkRole); err != nil {
				slog.Error("invite: [4/6] Clerk AddOrgMembership FAILED",
					"clerk_user_id", clerkUserID, "clerk_org_id", *clerkOrgID, "error", err)
				return nil, fmt.Errorf("failed to add to organization: %w", err)
			}
			slog.Info("invite: [4/6] Clerk org membership added OK",
				"clerk_user_id", clerkUserID, "clerk_org_id", *clerkOrgID)
		} else {
			slog.Warn("invite: [4/6] SKIPPED org membership — org has no Clerk external_id",
				"org_id", inv.OrgID, "clerk_org_id", clerkOrgStr)
		}
	} else {
		slog.Warn("invite: [3-4/6] SKIPPED Clerk user creation — ClerkClient is nil (CLERK_SECRET_KEY not set)")
	}

	// Upsert local DB user (handles race with Clerk webhook)
	userID := uuid.New()
	upsertQuery := `
		INSERT INTO users (id, org_id, email, name, role, external_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (external_id) DO UPDATE SET
			name = EXCLUDED.name,
			role = EXCLUDED.role
		RETURNING id
	`
	var externalID *string
	if clerkUserID != "" {
		externalID = &clerkUserID
	}

	slog.Info("invite: [5/6] upserting local DB user",
		"email", inv.Email, "org_id", inv.OrgID, "role", inv.Role,
		"clerk_user_id", clerkUserID, "external_id_nil", externalID == nil)

	err = s.db.QueryRow(ctx, upsertQuery,
		userID, inv.OrgID, inv.Email, name, string(inv.Role), externalID,
	).Scan(&userID)
	if err != nil {
		slog.Error("invite: [5/6] DB upsert FAILED", "error", err, "email", inv.Email)
		return nil, fmt.Errorf("failed to create local user: %w", err)
	}

	slog.Info("invite: [5/6] local DB user upserted OK",
		"user_id", userID, "email", inv.Email, "clerk_user_id", clerkUserID)

	// Mark invitation as accepted LAST (so failures above leave invite retryable)
	slog.Info("invite: [6/6] marking invitation accepted", "invite_id", inv.ID)
	acceptQuery := `UPDATE invitations SET accepted_at = NOW() WHERE id = $1`
	_, err = s.db.Exec(ctx, acceptQuery, inv.ID)
	if err != nil {
		slog.Error("invite: [6/6] mark accepted FAILED", "invite_id", inv.ID, "error", err)
		return nil, fmt.Errorf("failed to mark invitation accepted: %w", err)
	}

	slog.Info("invite: accept flow COMPLETED",
		"user_id", userID, "email", inv.Email, "clerk_user_id", clerkUserID)

	return &models.User{
		ID:        userID,
		OrgID:     inv.OrgID,
		Email:     inv.Email,
		Name:      name,
		Role:      models.UserRole(inv.Role),
		CreatedAt: time.Now(),
	}, nil
}

// splitName splits a full name into first and last name.
func splitName(name string) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(name), " ", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
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
