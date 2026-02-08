package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
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

// ---- Portal Dashboard API methods ----

// PortalProject is a minimal project representation for portal contacts.
type PortalProject struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Address string    `json:"address"`
	Status  string    `json:"status"`
}

// GetContactProjects returns all projects a contact is assigned to.
func (s *PortalService) GetContactProjects(ctx context.Context, contactID uuid.UUID) ([]PortalProject, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.address, p.status
		FROM projects p
		JOIN project_assignments pa ON pa.project_id = p.id
		WHERE pa.contact_id = $1
		ORDER BY p.name
	`
	rows, err := s.db.Query(ctx, query, contactID)
	if err != nil {
		return nil, fmt.Errorf("get contact projects: %w", err)
	}
	defer rows.Close()

	var projects []PortalProject
	for rows.Next() {
		var p PortalProject
		if err := rows.Scan(&p.ID, &p.Name, &p.Address, &p.Status); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}
	if projects == nil {
		projects = []PortalProject{}
	}
	return projects, nil
}

// GetContactProjectTasks returns tasks within a specific project scoped to the contact's assignments.
func (s *PortalService) GetContactProjectTasks(ctx context.Context, contactID, projectID uuid.UUID) ([]models.ProjectTask, error) {
	query := `
		SELECT pt.id, pt.project_id, pt.wbs_code, pt.name, pt.status, pt.planned_start, pt.planned_end
		FROM project_tasks pt
		JOIN project_assignments pa ON pt.project_id = pa.project_id
		WHERE pa.contact_id = $1
		  AND pt.project_id = $2
		  AND pt.wbs_code LIKE pa.wbs_phase_id || '.%'
		ORDER BY pt.planned_start ASC NULLS LAST
	`
	rows, err := s.db.Query(ctx, query, contactID, projectID)
	if err != nil {
		return nil, fmt.Errorf("get contact project tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.ProjectTask
	for rows.Next() {
		var t models.ProjectTask
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.WBSCode, &t.Name, &t.Status, &t.PlannedStart, &t.PlannedEnd); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	if tasks == nil {
		tasks = []models.ProjectTask{}
	}
	return tasks, nil
}

// ContactHasProjectAccess checks whether a contact is assigned to a given project.
func (s *PortalService) ContactHasProjectAccess(ctx context.Context, contactID, projectID uuid.UUID) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM project_assignments WHERE contact_id = $1 AND project_id = $2)`,
		contactID, projectID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check project access: %w", err)
	}
	return exists, nil
}

// ---- Dependency Schedule ----

// DependencyTask represents a task with trade name for dependency views.
type DependencyTask struct {
	ID           uuid.UUID  `json:"id"`
	WBSCode      string     `json:"wbs_code"`
	Name         string     `json:"name"`
	Status       string     `json:"status"`
	PlannedStart *time.Time `json:"planned_start,omitempty"`
	PlannedEnd   *time.Time `json:"planned_end,omitempty"`
	TradeName    string     `json:"trade_name,omitempty"`
	Relationship string     `json:"relationship,omitempty"` // "blocks_me" or "blocked_by_me"
}

// DependencySchedule contains a contact's tasks and their predecessor/successor relationships.
type DependencySchedule struct {
	MyTasks      []DependencyTask `json:"my_tasks"`
	Predecessors []DependencyTask `json:"predecessors"`
	Successors   []DependencyTask `json:"successors"`
}

// GetDependencySchedule returns the contact's tasks plus immediate predecessors/successors.
func (s *PortalService) GetDependencySchedule(ctx context.Context, contactID, projectID uuid.UUID) (*DependencySchedule, error) {
	// 1. Get contact's tasks for this project
	myTasksQuery := `
		SELECT pt.id, pt.wbs_code, pt.name, pt.status, pt.planned_start, pt.planned_end
		FROM project_tasks pt
		JOIN project_assignments pa ON pt.project_id = pa.project_id
		WHERE pa.contact_id = $1
		  AND pt.project_id = $2
		  AND pt.wbs_code LIKE pa.wbs_phase_id || '.%'
		ORDER BY pt.planned_start ASC NULLS LAST
	`
	rows, err := s.db.Query(ctx, myTasksQuery, contactID, projectID)
	if err != nil {
		return nil, fmt.Errorf("get dependency schedule tasks: %w", err)
	}
	defer rows.Close()

	var myTasks []DependencyTask
	var taskIDs []uuid.UUID
	for rows.Next() {
		var t DependencyTask
		if err := rows.Scan(&t.ID, &t.WBSCode, &t.Name, &t.Status, &t.PlannedStart, &t.PlannedEnd); err != nil {
			return nil, fmt.Errorf("scan dependency task: %w", err)
		}
		myTasks = append(myTasks, t)
		taskIDs = append(taskIDs, t.ID)
	}

	if myTasks == nil {
		return &DependencySchedule{
			MyTasks:      []DependencyTask{},
			Predecessors: []DependencyTask{},
			Successors:   []DependencyTask{},
		}, nil
	}

	// 2. Get predecessors (tasks that block my tasks)
	predQuery := `
		SELECT DISTINCT pt.id, pt.wbs_code, pt.name, pt.status, pt.planned_start, pt.planned_end,
		       COALESCE(wp.phase_name, '') AS trade_name
		FROM task_dependencies td
		JOIN project_tasks pt ON pt.id = td.predecessor_id
		LEFT JOIN wbs_phases wp ON wp.wbs_code = SUBSTRING(pt.wbs_code FROM '^[0-9]+\.[0-9]+')
		WHERE td.successor_id = ANY($1)
		  AND td.project_id = $2
		  AND pt.id != ALL($1)
	`
	predRows, err := s.db.Query(ctx, predQuery, taskIDs, projectID)
	if err != nil {
		return nil, fmt.Errorf("get predecessors: %w", err)
	}
	defer predRows.Close()

	var predecessors []DependencyTask
	for predRows.Next() {
		var t DependencyTask
		if err := predRows.Scan(&t.ID, &t.WBSCode, &t.Name, &t.Status, &t.PlannedStart, &t.PlannedEnd, &t.TradeName); err != nil {
			return nil, fmt.Errorf("scan predecessor: %w", err)
		}
		t.Relationship = "blocks_me"
		predecessors = append(predecessors, t)
	}

	// 3. Get successors (tasks blocked by my tasks)
	succQuery := `
		SELECT DISTINCT pt.id, pt.wbs_code, pt.name, pt.status, pt.planned_start, pt.planned_end,
		       COALESCE(wp.phase_name, '') AS trade_name
		FROM task_dependencies td
		JOIN project_tasks pt ON pt.id = td.successor_id
		LEFT JOIN wbs_phases wp ON wp.wbs_code = SUBSTRING(pt.wbs_code FROM '^[0-9]+\.[0-9]+')
		WHERE td.predecessor_id = ANY($1)
		  AND td.project_id = $2
		  AND pt.id != ALL($1)
	`
	succRows, err := s.db.Query(ctx, succQuery, taskIDs, projectID)
	if err != nil {
		return nil, fmt.Errorf("get successors: %w", err)
	}
	defer succRows.Close()

	var successors []DependencyTask
	for succRows.Next() {
		var t DependencyTask
		if err := succRows.Scan(&t.ID, &t.WBSCode, &t.Name, &t.Status, &t.PlannedStart, &t.PlannedEnd, &t.TradeName); err != nil {
			return nil, fmt.Errorf("scan successor: %w", err)
		}
		t.Relationship = "blocked_by_me"
		successors = append(successors, t)
	}

	if predecessors == nil {
		predecessors = []DependencyTask{}
	}
	if successors == nil {
		successors = []DependencyTask{}
	}

	return &DependencySchedule{
		MyTasks:      myTasks,
		Predecessors: predecessors,
		Successors:   successors,
	}, nil
}

// ---- Portal Messaging ----

// PortalMessage is a minimal message representation for portal threads.
type PortalMessage struct {
	ID          uuid.UUID  `json:"id"`
	Content     string     `json:"content"`
	SenderName  string     `json:"sender_name"`
	SenderType  string     `json:"sender_type"` // "user" or "contact"
	CreatedAt   time.Time  `json:"created_at"`
}

// GetOrCreatePortalThread finds or creates a portal-type thread for a contact within a project.
func (s *PortalService) GetOrCreatePortalThread(ctx context.Context, contactID, projectID uuid.UUID, contactName string) (uuid.UUID, error) {
	// Try to find existing portal thread for this contact
	var threadID uuid.UUID
	err := s.db.QueryRow(ctx,
		`SELECT id FROM threads
		 WHERE project_id = $1 AND thread_type = 'portal' AND created_by = $2
		 LIMIT 1`,
		projectID, contactID,
	).Scan(&threadID)
	if err == nil {
		return threadID, nil
	}
	if err != pgx.ErrNoRows {
		return uuid.Nil, fmt.Errorf("lookup portal thread: %w", err)
	}

	// Create new portal thread
	threadID = uuid.New()
	title := fmt.Sprintf("Portal: %s", contactName)
	_, err = s.db.Exec(ctx,
		`INSERT INTO threads (id, project_id, title, is_general, thread_type, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, false, 'portal', $4, NOW(), NOW())`,
		threadID, projectID, title, contactID,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create portal thread: %w", err)
	}

	slog.Info("portal: created portal thread", "thread_id", threadID, "contact_id", contactID, "project_id", projectID)
	return threadID, nil
}

// GetPortalMessages returns messages from a contact's portal thread.
func (s *PortalService) GetPortalMessages(ctx context.Context, contactID, projectID uuid.UUID, limit int) ([]PortalMessage, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	query := `
		SELECT cm.id, cm.content, cm.created_at,
		       CASE
		           WHEN cm.user_id IS NOT NULL THEN COALESCE(u.name, 'Team')
		           WHEN cm.contact_id IS NOT NULL THEN COALESCE(c.name, 'Contact')
		           ELSE 'System'
		       END AS sender_name,
		       CASE
		           WHEN cm.user_id IS NOT NULL THEN 'user'
		           WHEN cm.contact_id IS NOT NULL THEN 'contact'
		           ELSE 'system'
		       END AS sender_type
		FROM chat_messages cm
		JOIN threads t ON t.id = cm.thread_id
		LEFT JOIN users u ON u.id = cm.user_id
		LEFT JOIN contacts c ON c.id = cm.contact_id
		WHERE t.project_id = $1
		  AND t.thread_type = 'portal'
		  AND t.created_by = $2
		ORDER BY cm.created_at ASC
		LIMIT $3
	`
	rows, err := s.db.Query(ctx, query, projectID, contactID, limit)
	if err != nil {
		return nil, fmt.Errorf("get portal messages: %w", err)
	}
	defer rows.Close()

	var messages []PortalMessage
	for rows.Next() {
		var m PortalMessage
		if err := rows.Scan(&m.ID, &m.Content, &m.CreatedAt, &m.SenderName, &m.SenderType); err != nil {
			return nil, fmt.Errorf("scan portal message: %w", err)
		}
		messages = append(messages, m)
	}
	if messages == nil {
		messages = []PortalMessage{}
	}
	return messages, nil
}

// SendPortalMessage persists a message from a portal contact.
func (s *PortalService) SendPortalMessage(ctx context.Context, contactID, projectID uuid.UUID, contactName, content string) (*PortalMessage, error) {
	threadID, err := s.GetOrCreatePortalThread(ctx, contactID, projectID, contactName)
	if err != nil {
		return nil, fmt.Errorf("get portal thread: %w", err)
	}

	msgID := uuid.New()
	now := time.Now().UTC()

	_, err = s.db.Exec(ctx,
		`INSERT INTO chat_messages (id, project_id, thread_id, contact_id, role, content, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		msgID, projectID, threadID, contactID, types.ChatRoleUser, content, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert portal message: %w", err)
	}

	return &PortalMessage{
		ID:         msgID,
		Content:    content,
		SenderName: contactName,
		SenderType: "contact",
		CreatedAt:  now,
	}, nil
}

// ---- Portal Documents ----

// PortalDocument is a minimal document representation for portal contacts.
type PortalDocument struct {
	ID               uuid.UUID `json:"id"`
	Filename         string    `json:"filename"`
	MimeType         string    `json:"mime_type"`
	FileSizeBytes    int64     `json:"file_size_bytes"`
	ProcessingStatus string    `json:"processing_status"`
	UploadedAt       time.Time `json:"uploaded_at"`
}

// ListContactDocuments returns documents uploaded by a contact for a project.
func (s *PortalService) ListContactDocuments(ctx context.Context, contactID, projectID uuid.UUID) ([]PortalDocument, error) {
	query := `
		SELECT id, filename, mime_type, file_size_bytes, processing_status, uploaded_at
		FROM documents
		WHERE project_id = $1 AND uploaded_by = $2
		ORDER BY uploaded_at DESC
	`
	rows, err := s.db.Query(ctx, query, projectID, contactID)
	if err != nil {
		return nil, fmt.Errorf("list contact documents: %w", err)
	}
	defer rows.Close()

	var docs []PortalDocument
	for rows.Next() {
		var d PortalDocument
		if err := rows.Scan(&d.ID, &d.Filename, &d.MimeType, &d.FileSizeBytes, &d.ProcessingStatus, &d.UploadedAt); err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}
		docs = append(docs, d)
	}
	if docs == nil {
		docs = []PortalDocument{}
	}
	return docs, nil
}

// CreateDocument inserts a document record for a portal upload.
func (s *PortalService) CreateDocument(ctx context.Context, projectID, uploadedBy uuid.UUID, filename, storagePath, mimeType string, fileSizeBytes int64) (*PortalDocument, error) {
	docID := uuid.New()
	now := time.Now().UTC()

	_, err := s.db.Exec(ctx,
		`INSERT INTO documents (id, project_id, filename, storage_path, mime_type, file_size_bytes, processing_status, uploaded_by, uploaded_at)
		 VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7, $8)`,
		docID, projectID, filename, storagePath, mimeType, fileSizeBytes, uploadedBy, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create document: %w", err)
	}

	return &PortalDocument{
		ID:               docID,
		Filename:         filename,
		MimeType:         mimeType,
		FileSizeBytes:    fileSizeBytes,
		ProcessingStatus: "pending",
		UploadedAt:       now,
	}, nil
}

// ---- Portal Invoices (Subcontractor) ----

// PortalInvoice is a minimal invoice representation for portal contacts.
type PortalInvoice struct {
	ID            uuid.UUID  `json:"id"`
	VendorName    string     `json:"vendor_name"`
	AmountCents   int64      `json:"amount_cents"`
	Status        string     `json:"status"`
	InvoiceNumber *string    `json:"invoice_number,omitempty"`
	InvoiceDate   *time.Time `json:"invoice_date,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ListContactInvoices returns invoices for tasks in the contact's assigned phases.
func (s *PortalService) ListContactInvoices(ctx context.Context, contactID, projectID uuid.UUID) ([]PortalInvoice, error) {
	query := `
		SELECT i.id, i.vendor_name, i.amount_cents, i.status, i.invoice_number, i.invoice_date, i.created_at
		FROM invoices i
		JOIN project_assignments pa ON i.project_id = pa.project_id
		WHERE pa.contact_id = $1
		  AND i.project_id = $2
		  AND i.detected_wbs_code LIKE pa.wbs_phase_id || '.%'
		ORDER BY i.created_at DESC
	`
	rows, err := s.db.Query(ctx, query, contactID, projectID)
	if err != nil {
		return nil, fmt.Errorf("list contact invoices: %w", err)
	}
	defer rows.Close()

	var invoices []PortalInvoice
	for rows.Next() {
		var inv PortalInvoice
		if err := rows.Scan(&inv.ID, &inv.VendorName, &inv.AmountCents, &inv.Status, &inv.InvoiceNumber, &inv.InvoiceDate, &inv.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan invoice: %w", err)
		}
		invoices = append(invoices, inv)
	}
	if invoices == nil {
		invoices = []PortalInvoice{}
	}
	return invoices, nil
}
