package service

import (
	"context"
	"fmt"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ThreadService handles CRUD operations for conversation threads.
type ThreadService struct {
	db *pgxpool.Pool
}

// NewThreadService creates a new ThreadService.
func NewThreadService(db *pgxpool.Pool) *ThreadService {
	return &ThreadService{db: db}
}

// CreateThread creates a new non-general thread in a project.
// Enforces multi-tenancy via JOIN on projects.org_id.
func (s *ThreadService) CreateThread(ctx context.Context, projectID, orgID, userID uuid.UUID, title string) (*models.Thread, error) {
	thread := &models.Thread{
		ID:        uuid.New(),
		ProjectID: projectID,
		Title:     title,
		IsGeneral: false,
		CreatedBy: &userID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	query := `
		INSERT INTO threads (id, project_id, title, is_general, created_by, created_at, updated_at)
		SELECT $1, p.id, $3, false, $4, $5, $6
		FROM projects p WHERE p.id = $2 AND p.org_id = $7
		RETURNING id
	`
	var id uuid.UUID
	err := s.db.QueryRow(ctx, query,
		thread.ID, projectID, title, userID, thread.CreatedAt, thread.UpdatedAt, orgID,
	).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.ErrNotFound
		}
		return nil, fmt.Errorf("create thread: %w", err)
	}

	return thread, nil
}

// CreateGeneralThread creates or returns the General thread for a project.
// Idempotent via ON CONFLICT on the partial unique index.
func (s *ThreadService) CreateGeneralThread(ctx context.Context, projectID uuid.UUID) (*models.Thread, error) {
	now := time.Now().UTC()
	id := uuid.New()

	query := `
		INSERT INTO threads (id, project_id, title, is_general, created_at, updated_at)
		VALUES ($1, $2, 'General', true, $3, $4)
		ON CONFLICT (project_id) WHERE is_general = true
		DO UPDATE SET updated_at = $4
		RETURNING id, project_id, title, is_general, archived_at, created_by, created_at, updated_at
	`
	thread := &models.Thread{}
	err := s.db.QueryRow(ctx, query, id, projectID, now, now).Scan(
		&thread.ID, &thread.ProjectID, &thread.Title, &thread.IsGeneral,
		&thread.ArchivedAt, &thread.CreatedBy, &thread.CreatedAt, &thread.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create general thread: %w", err)
	}

	return thread, nil
}

// ListThreads returns all threads for a project. Excludes archived unless includeArchived is true.
// Enforces multi-tenancy via JOIN on projects.org_id.
func (s *ThreadService) ListThreads(ctx context.Context, projectID, orgID uuid.UUID, includeArchived bool) ([]models.Thread, error) {
	query := `
		SELECT t.id, t.project_id, t.title, t.is_general, t.archived_at, t.created_by, t.created_at, t.updated_at
		FROM threads t
		JOIN projects p ON p.id = t.project_id
		WHERE t.project_id = $1 AND p.org_id = $2
	`
	if !includeArchived {
		query += " AND t.archived_at IS NULL"
	}
	query += " ORDER BY t.is_general DESC, t.created_at ASC"

	rows, err := s.db.Query(ctx, query, projectID, orgID)
	if err != nil {
		return nil, fmt.Errorf("list threads: %w", err)
	}
	defer rows.Close()

	var threads []models.Thread
	for rows.Next() {
		var t models.Thread
		if err := rows.Scan(
			&t.ID, &t.ProjectID, &t.Title, &t.IsGeneral,
			&t.ArchivedAt, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan thread: %w", err)
		}
		threads = append(threads, t)
	}

	if threads == nil {
		threads = []models.Thread{}
	}

	return threads, nil
}

// GetThread returns a single thread by ID. Enforces multi-tenancy.
func (s *ThreadService) GetThread(ctx context.Context, threadID, projectID, orgID uuid.UUID) (*models.Thread, error) {
	query := `
		SELECT t.id, t.project_id, t.title, t.is_general, t.archived_at, t.created_by, t.created_at, t.updated_at
		FROM threads t
		JOIN projects p ON p.id = t.project_id
		WHERE t.id = $1 AND t.project_id = $2 AND p.org_id = $3
	`
	t := &models.Thread{}
	err := s.db.QueryRow(ctx, query, threadID, projectID, orgID).Scan(
		&t.ID, &t.ProjectID, &t.Title, &t.IsGeneral,
		&t.ArchivedAt, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.ErrNotFound
		}
		return nil, fmt.Errorf("get thread: %w", err)
	}

	return t, nil
}

// ArchiveThread soft-deletes a thread by setting archived_at.
// Rejects archival of General threads (403).
func (s *ThreadService) ArchiveThread(ctx context.Context, threadID, projectID, orgID uuid.UUID) error {
	// First check if it's a General thread
	thread, err := s.GetThread(ctx, threadID, projectID, orgID)
	if err != nil {
		return err
	}
	if thread.IsGeneral {
		return types.ErrForbidden
	}

	query := `
		UPDATE threads SET archived_at = $1, updated_at = $1
		WHERE id = $2 AND project_id = $3
	`
	_, err = s.db.Exec(ctx, query, time.Now().UTC(), threadID, projectID)
	if err != nil {
		return fmt.Errorf("archive thread: %w", err)
	}

	return nil
}

// UnarchiveThread restores an archived thread.
func (s *ThreadService) UnarchiveThread(ctx context.Context, threadID, projectID, orgID uuid.UUID) error {
	// Verify thread exists and belongs to this org
	_, err := s.GetThread(ctx, threadID, projectID, orgID)
	if err != nil {
		return err
	}

	query := `
		UPDATE threads SET archived_at = NULL, updated_at = $1
		WHERE id = $2 AND project_id = $3
	`
	_, err = s.db.Exec(ctx, query, time.Now().UTC(), threadID, projectID)
	if err != nil {
		return fmt.Errorf("unarchive thread: %w", err)
	}

	return nil
}

// GetOrCreateGeneralThread returns the General thread for a project,
// creating one if it doesn't exist. Used as fallback for messages without a thread.
func (s *ThreadService) GetOrCreateGeneralThread(ctx context.Context, projectID uuid.UUID) (*models.Thread, error) {
	// Try to find existing General thread
	query := `
		SELECT id, project_id, title, is_general, archived_at, created_by, created_at, updated_at
		FROM threads
		WHERE project_id = $1 AND is_general = true
	`
	t := &models.Thread{}
	err := s.db.QueryRow(ctx, query, projectID).Scan(
		&t.ID, &t.ProjectID, &t.Title, &t.IsGeneral,
		&t.ArchivedAt, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == nil {
		return t, nil
	}
	if err != pgx.ErrNoRows {
		return nil, fmt.Errorf("get general thread: %w", err)
	}

	// Create one
	return s.CreateGeneralThread(ctx, projectID)
}

// GetThreadMessages returns chat messages for a specific thread, ordered by created_at.
// Enforces multi-tenancy via JOIN on projects.org_id.
func (s *ThreadService) GetThreadMessages(ctx context.Context, threadID, projectID, orgID uuid.UUID, limit int) ([]models.ChatMessage, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	query := `
		SELECT cm.id, cm.project_id, cm.thread_id, cm.user_id, cm.role, cm.content, cm.tool_calls, cm.created_at
		FROM chat_messages cm
		JOIN threads t ON t.id = cm.thread_id
		JOIN projects p ON p.id = t.project_id
		WHERE cm.thread_id = $1 AND t.project_id = $2 AND p.org_id = $3
		ORDER BY cm.created_at ASC
		LIMIT $4
	`
	rows, err := s.db.Query(ctx, query, threadID, projectID, orgID, limit)
	if err != nil {
		return nil, fmt.Errorf("get thread messages: %w", err)
	}
	defer rows.Close()

	var messages []models.ChatMessage
	for rows.Next() {
		var m models.ChatMessage
		if err := rows.Scan(
			&m.ID, &m.ProjectID, &m.ThreadID, &m.UserID, &m.Role, &m.Content, &m.ToolCalls, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, m)
	}

	if messages == nil {
		messages = []models.ChatMessage{}
	}

	return messages, nil
}
