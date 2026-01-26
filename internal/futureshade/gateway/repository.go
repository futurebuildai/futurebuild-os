package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ExecutionStatus represents the status of a skill execution.
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "PENDING"
	StatusRunning   ExecutionStatus = "RUNNING"
	StatusCompleted ExecutionStatus = "COMPLETED"
	StatusFailed    ExecutionStatus = "FAILED"
)

// ExecutionLog represents a record in shadow_execution_logs.
type ExecutionLog struct {
	ID            uuid.UUID              `json:"id"`
	DecisionID    uuid.UUID              `json:"decision_id"`
	SkillID       string                 `json:"skill_id"`
	Parameters    map[string]any         `json:"parameters"`
	Status        ExecutionStatus        `json:"status"`
	ResultSummary *string                `json:"result_summary,omitempty"`
	ErrorMessage  *string                `json:"error_message,omitempty"`
	StartedAt     *time.Time             `json:"started_at,omitempty"`
	FinishedAt    *time.Time             `json:"finished_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

// Repository provides CRUD operations for shadow_execution_logs.
// See specs/FUTURESHADE_AGENTS_SPEC.md Section 4.4 (Gateway Repository)
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new gateway repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateExecutionLog creates a new execution log entry in PENDING state.
func (r *Repository) CreateExecutionLog(ctx context.Context, decisionID uuid.UUID, skillID string, params map[string]any) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx, `
		INSERT INTO shadow_execution_logs (decision_id, skill_id, parameters, status)
		VALUES ($1, $2, $3, 'PENDING')
		RETURNING id
	`, decisionID, skillID, params).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create execution log: %w", err)
	}
	return id, nil
}

// MarkRunning transitions an execution log to RUNNING state.
// Records the start time.
func (r *Repository) MarkRunning(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx, `
		UPDATE shadow_execution_logs
		SET status = 'RUNNING', started_at = NOW()
		WHERE id = $1 AND status = 'PENDING'
	`, id)
	if err != nil {
		return fmt.Errorf("mark running: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("execution log %s not found or not in PENDING state", id)
	}
	return nil
}

// UpdateExecutionStatus updates the status and result/error of an execution log.
// Used to mark COMPLETED or FAILED.
func (r *Repository) UpdateExecutionStatus(ctx context.Context, id uuid.UUID, status ExecutionStatus, resultSummary, errorMessage *string) error {
	result, err := r.db.Exec(ctx, `
		UPDATE shadow_execution_logs
		SET status = $2, result_summary = $3, error_message = $4, finished_at = NOW()
		WHERE id = $1
	`, id, status, resultSummary, errorMessage)
	if err != nil {
		return fmt.Errorf("update execution status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("execution log %s not found", id)
	}
	return nil
}

// GetExecutionLog retrieves an execution log by ID.
func (r *Repository) GetExecutionLog(ctx context.Context, id uuid.UUID) (*ExecutionLog, error) {
	var log ExecutionLog
	err := r.db.QueryRow(ctx, `
		SELECT id, decision_id, skill_id, parameters, status,
		       result_summary, error_message, started_at, finished_at, created_at
		FROM shadow_execution_logs WHERE id = $1
	`, id).Scan(
		&log.ID, &log.DecisionID, &log.SkillID, &log.Parameters, &log.Status,
		&log.ResultSummary, &log.ErrorMessage, &log.StartedAt, &log.FinishedAt, &log.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get execution log: %w", err)
	}
	return &log, nil
}

// ListByDecision retrieves all execution logs for a given decision.
func (r *Repository) ListByDecision(ctx context.Context, decisionID uuid.UUID) ([]ExecutionLog, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, decision_id, skill_id, parameters, status,
		       result_summary, error_message, started_at, finished_at, created_at
		FROM shadow_execution_logs
		WHERE decision_id = $1
		ORDER BY created_at ASC
	`, decisionID)
	if err != nil {
		return nil, fmt.Errorf("list by decision: %w", err)
	}
	defer rows.Close()

	var logs []ExecutionLog
	for rows.Next() {
		var log ExecutionLog
		err := rows.Scan(
			&log.ID, &log.DecisionID, &log.SkillID, &log.Parameters, &log.Status,
			&log.ResultSummary, &log.ErrorMessage, &log.StartedAt, &log.FinishedAt, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan execution log: %w", err)
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

// GetStatus retrieves just the status of an execution log.
// Useful for idempotency checks.
func (r *Repository) GetStatus(ctx context.Context, id uuid.UUID) (ExecutionStatus, error) {
	var status ExecutionStatus
	err := r.db.QueryRow(ctx, `
		SELECT status FROM shadow_execution_logs WHERE id = $1
	`, id).Scan(&status)
	if err != nil {
		return "", fmt.Errorf("get status: %w", err)
	}
	return status, nil
}
