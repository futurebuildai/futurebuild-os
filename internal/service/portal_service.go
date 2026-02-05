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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ActionType represents the type of portal action.
type ActionType string

const (
	ActionTypeStatusUpdate ActionType = "status_update"
	ActionTypePhotoUpload  ActionType = "photo_upload"
	ActionTypeView         ActionType = "view"
)

// ActionToken represents a one-time portal action token.
type ActionToken struct {
	ID         uuid.UUID  `json:"id"`
	TokenHash  string     `json:"-"`
	ContactID  uuid.UUID  `json:"contact_id"`
	ProjectID  uuid.UUID  `json:"project_id"`
	TaskID     uuid.UUID  `json:"task_id"`
	ActionType ActionType `json:"action_type"`
	ExpiresAt  time.Time  `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UsedAt     *time.Time `json:"used_at,omitempty"`
	Used       bool       `json:"used"`
}

// ActionTokenContext contains the full context for a portal action.
type ActionTokenContext struct {
	Token   *ActionToken        `json:"token"`
	Contact *models.Contact     `json:"contact"`
	Project *models.Project     `json:"project"`
	Task    *models.ProjectTask `json:"task"`
}

// PortalService handles portal action token operations.
// See LAUNCH_PLAN.md P2: Field Portal (Mobile).
type PortalService struct {
	db                  *pgxpool.Pool
	notificationService types.NotificationService
	baseURL             string
}

// NewPortalService creates a new PortalService.
func NewPortalService(db *pgxpool.Pool, notificationService types.NotificationService, baseURL string) *PortalService {
	return &PortalService{
		db:                  db,
		notificationService: notificationService,
		baseURL:             baseURL,
	}
}

// GenerateToken creates a cryptographically secure 32-byte random string.
func (s *PortalService) GenerateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashToken returns the SHA-256 hash of a plaintext token.
func (s *PortalService) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CreateActionToken creates a one-time action token for a contact/task pair.
// Returns the raw (unhashed) token that should be sent to the contact.
func (s *PortalService) CreateActionToken(
	ctx context.Context,
	contactID, projectID, taskID uuid.UUID,
	actionType ActionType,
) (string, error) {
	rawToken, err := s.GenerateToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	tokenHash := s.HashToken(rawToken)
	expiresAt := time.Now().UTC().Add(48 * time.Hour) // 48 hour expiry

	query := `
		INSERT INTO portal_action_tokens (token_hash, contact_id, project_id, task_id, action_type, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = s.db.Exec(ctx, query, tokenHash, contactID, projectID, taskID, actionType, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to store action token: %w", err)
	}

	return rawToken, nil
}

// VerifyActionToken validates and returns the action token context.
// Does NOT mark the token as used - call UseActionToken after action is completed.
func (s *PortalService) VerifyActionToken(ctx context.Context, rawToken string) (*ActionTokenContext, error) {
	tokenHash := s.HashToken(rawToken)

	// Query token first
	tokenQuery := `
		SELECT id, contact_id, project_id, task_id, action_type, expires_at, created_at, used_at, used
		FROM portal_action_tokens
		WHERE token_hash = $1
	`

	var token ActionToken
	err := s.db.QueryRow(ctx, tokenQuery, tokenHash).Scan(
		&token.ID, &token.ContactID, &token.ProjectID, &token.TaskID, &token.ActionType,
		&token.ExpiresAt, &token.CreatedAt, &token.UsedAt, &token.Used,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("invalid or expired token")
		}
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	// Validate token status
	if token.Used {
		return nil, fmt.Errorf("token already used")
	}
	if time.Now().After(token.ExpiresAt) {
		return nil, fmt.Errorf("token expired")
	}

	// Query contact
	var contact models.Contact
	contactQuery := `
		SELECT id, org_id, name, company, phone, email, role, contact_preference, created_at
		FROM contacts WHERE id = $1
	`
	err = s.db.QueryRow(ctx, contactQuery, token.ContactID).Scan(
		&contact.ID, &contact.OrgID, &contact.Name, &contact.Company, &contact.Phone,
		&contact.Email, &contact.Role, &contact.ContactPreference, &contact.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	// Query project
	var project models.Project
	projectQuery := `SELECT id, org_id, name, address, status FROM projects WHERE id = $1`
	err = s.db.QueryRow(ctx, projectQuery, token.ProjectID).Scan(
		&project.ID, &project.OrgID, &project.Name, &project.Address, &project.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Query task
	var task models.ProjectTask
	taskQuery := `
		SELECT id, project_id, wbs_code, name, status, planned_start, planned_end
		FROM project_tasks WHERE id = $1
	`
	err = s.db.QueryRow(ctx, taskQuery, token.TaskID).Scan(
		&task.ID, &task.ProjectID, &task.WBSCode, &task.Name, &task.Status,
		&task.PlannedStart, &task.PlannedEnd,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &ActionTokenContext{
		Token:   &token,
		Contact: &contact,
		Project: &project,
		Task:    &task,
	}, nil
}

// UseActionToken marks an action token as used.
func (s *PortalService) UseActionToken(ctx context.Context, tokenID uuid.UUID) error {
	query := `
		UPDATE portal_action_tokens
		SET used = true, used_at = NOW()
		WHERE id = $1 AND used = false
	`
	result, err := s.db.Exec(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("token already used or not found")
	}
	return nil
}

// SendActionLink generates a token and sends an SMS with the one-time link.
func (s *PortalService) SendActionLink(
	ctx context.Context,
	contactID, projectID, taskID uuid.UUID,
	actionType ActionType,
) error {
	// Look up contact phone number
	var phone *string
	var contactName string
	err := s.db.QueryRow(ctx, "SELECT phone, name FROM contacts WHERE id = $1", contactID).Scan(&phone, &contactName)
	if err != nil {
		return fmt.Errorf("failed to look up contact: %w", err)
	}
	if phone == nil || *phone == "" {
		return fmt.Errorf("contact has no phone number")
	}

	// Look up task name
	var taskName string
	err = s.db.QueryRow(ctx, "SELECT name FROM project_tasks WHERE id = $1", taskID).Scan(&taskName)
	if err != nil {
		return fmt.Errorf("failed to look up task: %w", err)
	}

	// Generate token
	rawToken, err := s.CreateActionToken(ctx, contactID, projectID, taskID, actionType)
	if err != nil {
		return fmt.Errorf("failed to create action token: %w", err)
	}

	// Build link
	actionLink := fmt.Sprintf("%s/portal/action/%s", s.baseURL, rawToken)

	// Build message based on action type
	var message string
	switch actionType {
	case ActionTypeStatusUpdate:
		message = fmt.Sprintf("FutureBuild: Update needed for \"%s\". Tap to respond: %s", taskName, actionLink)
	case ActionTypePhotoUpload:
		message = fmt.Sprintf("FutureBuild: Photo needed for \"%s\". Tap to upload: %s", taskName, actionLink)
	default:
		message = fmt.Sprintf("FutureBuild: View task \"%s\": %s", taskName, actionLink)
	}

	// Send SMS
	if err := s.notificationService.SendSMS(*phone, message); err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	return nil
}

// UpdateTaskStatus updates a task's status via portal action.
func (s *PortalService) UpdateTaskStatus(ctx context.Context, taskID, projectID uuid.UUID, status types.TaskStatus) error {
	query := `
		UPDATE project_tasks
		SET status = $1
		WHERE id = $2 AND project_id = $3
	`
	result, err := s.db.Exec(ctx, query, status, taskID, projectID)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

// GetContactTasks returns all tasks assigned to a contact across all projects.
// Tasks are scoped by the contact's WBS phase assignments (e.g. phase "5.3" matches tasks "5.3.1", "5.3.2").
func (s *PortalService) GetContactTasks(ctx context.Context, contactID uuid.UUID) ([]models.ProjectTask, error) {
	query := `
		SELECT pt.id, pt.project_id, pt.wbs_code, pt.name, pt.status, pt.planned_start, pt.planned_end
		FROM project_tasks pt
		JOIN project_assignments pa ON pt.project_id = pa.project_id
		WHERE pa.contact_id = $1
		  AND pt.wbs_code LIKE pa.wbs_phase_id || '.%'
		ORDER BY pt.planned_start ASC NULLS LAST
	`

	rows, err := s.db.Query(ctx, query, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.ProjectTask
	for rows.Next() {
		var task models.ProjectTask
		err := rows.Scan(
			&task.ID, &task.ProjectID, &task.WBSCode, &task.Name, &task.Status,
			&task.PlannedStart, &task.PlannedEnd,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
