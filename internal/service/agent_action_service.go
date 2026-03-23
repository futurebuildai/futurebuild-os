package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AgentActionService manages agent pending actions (human-in-the-loop approval).
type AgentActionService struct {
	db          *pgxpool.Pool
	feedService *FeedService
	toolRunner  ActionToolRunner
}

// ActionToolRunner executes tool calls from approved agent actions.
// Satisfied by the tools.Registry.Execute method wrapped with scope injection.
type ActionToolRunner interface {
	ExecuteAction(ctx context.Context, actionType string, payload json.RawMessage, scope ActionScope) (string, error)
}

// ActionScope provides context for tool execution during action approval.
type ActionScope struct {
	ProjectID uuid.UUID
	OrgID     uuid.UUID
	UserID    uuid.UUID
}

// NewAgentActionService creates an AgentActionService.
func NewAgentActionService(db *pgxpool.Pool, feedService *FeedService) *AgentActionService {
	return &AgentActionService{
		db:          db,
		feedService: feedService,
	}
}

// WithToolRunner sets the tool runner for executing approved actions.
func (s *AgentActionService) WithToolRunner(runner ActionToolRunner) *AgentActionService {
	s.toolRunner = runner
	return s
}

// CreatePendingAction inserts a new pending action tied to a feed card.
func (s *AgentActionService) CreatePendingAction(ctx context.Context, action *models.AgentPendingAction) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO agent_pending_actions (
			org_id, project_id, feed_card_id, agent_source,
			action_type, action_payload, status, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7)
		RETURNING id, created_at
	`,
		action.OrgID, action.ProjectID, action.FeedCardID, action.AgentSource,
		action.ActionType, action.ActionPayload, action.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("create pending action: %w", err)
	}
	return nil
}

// ListPending returns all pending agent actions for an organization.
func (s *AgentActionService) ListPending(ctx context.Context, orgID uuid.UUID) ([]models.AgentPendingAction, error) {
	rows, err := s.db.Query(ctx, `
		SELECT
			apa.id, apa.org_id, apa.project_id, apa.feed_card_id,
			apa.agent_source, apa.action_type, apa.action_payload,
			apa.status, apa.expires_at, apa.created_at,
			fc.headline,
			p.name AS project_name
		FROM agent_pending_actions apa
		JOIN feed_cards fc ON fc.id = apa.feed_card_id
		JOIN projects p ON p.id = apa.project_id
		WHERE apa.org_id = $1
			AND apa.status = 'pending'
			AND (apa.expires_at IS NULL OR apa.expires_at > NOW())
		ORDER BY apa.created_at DESC
		LIMIT 100
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("list pending actions: %w", err)
	}
	defer rows.Close()

	var actions []models.AgentPendingAction
	for rows.Next() {
		var a models.AgentPendingAction
		if err := rows.Scan(
			&a.ID, &a.OrgID, &a.ProjectID, &a.FeedCardID,
			&a.AgentSource, &a.ActionType, &a.ActionPayload,
			&a.Status, &a.ExpiresAt, &a.CreatedAt,
			&a.Headline, &a.ProjectName,
		); err != nil {
			return nil, fmt.Errorf("scan pending action: %w", err)
		}
		actions = append(actions, a)
	}

	if actions == nil {
		actions = []models.AgentPendingAction{}
	}
	return actions, rows.Err()
}

// ApproveAction approves a pending action and executes the stored tool call.
func (s *AgentActionService) ApproveAction(ctx context.Context, orgID uuid.UUID, actionID uuid.UUID, approverID uuid.UUID) (*ApprovalResult, error) {
	// 1. Fetch the pending action
	action, err := s.getPendingAction(ctx, orgID, actionID)
	if err != nil {
		return nil, err
	}

	// 2. Verify it's still pending and not expired
	if action.Status != models.PendingActionStatusPending {
		return nil, fmt.Errorf("action is no longer pending (status: %s)", action.Status)
	}
	if action.ExpiresAt != nil && action.ExpiresAt.Before(time.Now()) {
		// Auto-expire
		_ = s.updateStatus(ctx, actionID, models.PendingActionStatusExpired, nil, nil)
		return nil, fmt.Errorf("action has expired")
	}

	// 3. Mark as approved
	now := time.Now()
	if err := s.updateStatus(ctx, actionID, models.PendingActionStatusApproved, &approverID, &now); err != nil {
		return nil, fmt.Errorf("mark approved: %w", err)
	}

	// 4. Execute the action via tool runner
	result := &ApprovalResult{
		ActionID:   actionID,
		ActionType: action.ActionType,
		Status:     "approved",
	}

	if s.toolRunner != nil {
		scope := ActionScope{
			ProjectID: action.ProjectID,
			OrgID:     action.OrgID,
			UserID:    approverID,
		}
		output, err := s.toolRunner.ExecuteAction(ctx, action.ActionType, action.ActionPayload, scope)
		if err != nil {
			slog.Error("agent action execution failed after approval",
				"action_id", actionID,
				"action_type", action.ActionType,
				"error", err,
			)
			result.ExecutionError = err.Error()
		} else {
			result.ExecutionOutput = output
		}
	} else {
		slog.Warn("agent action approved but no tool runner configured", "action_id", actionID)
		result.ExecutionOutput = "Action approved (tool runner not configured)"
	}

	// 5. Dismiss the feed card
	if err := s.feedService.DismissCard(ctx, orgID, action.FeedCardID); err != nil {
		slog.Error("failed to dismiss feed card after approval", "card_id", action.FeedCardID, "error", err)
	}

	return result, nil
}

// RejectAction rejects a pending action with an optional reason.
func (s *AgentActionService) RejectAction(ctx context.Context, orgID uuid.UUID, actionID uuid.UUID, userID uuid.UUID, reason string) error {
	action, err := s.getPendingAction(ctx, orgID, actionID)
	if err != nil {
		return err
	}

	if action.Status != models.PendingActionStatusPending {
		return fmt.Errorf("action is no longer pending (status: %s)", action.Status)
	}

	now := time.Now()
	if err := s.updateStatusWithReason(ctx, actionID, models.PendingActionStatusRejected, &userID, &now, reason); err != nil {
		return fmt.Errorf("mark rejected: %w", err)
	}

	// Dismiss the feed card
	if err := s.feedService.DismissCard(ctx, orgID, action.FeedCardID); err != nil {
		slog.Error("failed to dismiss feed card after rejection", "card_id", action.FeedCardID, "error", err)
	}

	return nil
}

// ExpireStaleActions marks expired pending actions. Called periodically by a background worker.
func (s *AgentActionService) ExpireStaleActions(ctx context.Context) (int, error) {
	result, err := s.db.Exec(ctx, `
		UPDATE agent_pending_actions
		SET status = 'expired'
		WHERE status = 'pending'
			AND expires_at IS NOT NULL
			AND expires_at < NOW()
	`)
	if err != nil {
		return 0, fmt.Errorf("expire stale actions: %w", err)
	}
	return int(result.RowsAffected()), nil
}

// GetPendingByCardID retrieves a pending action by its associated feed card ID.
func (s *AgentActionService) GetPendingByCardID(ctx context.Context, orgID uuid.UUID, cardID uuid.UUID) (*models.AgentPendingAction, error) {
	var a models.AgentPendingAction
	err := s.db.QueryRow(ctx, `
		SELECT id, org_id, project_id, feed_card_id, agent_source,
			action_type, action_payload, status, approved_by, approved_at,
			rejection_reason, expires_at, created_at
		FROM agent_pending_actions
		WHERE feed_card_id = $1 AND org_id = $2
	`, cardID, orgID).Scan(
		&a.ID, &a.OrgID, &a.ProjectID, &a.FeedCardID, &a.AgentSource,
		&a.ActionType, &a.ActionPayload, &a.Status, &a.ApprovedBy, &a.ApprovedAt,
		&a.RejectionReason, &a.ExpiresAt, &a.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No pending action for this card
		}
		return nil, fmt.Errorf("get pending by card: %w", err)
	}
	return &a, nil
}

// ApprovalResult is the outcome of approving an agent action.
type ApprovalResult struct {
	ActionID        uuid.UUID `json:"action_id"`
	ActionType      string    `json:"action_type"`
	Status          string    `json:"status"`
	ExecutionOutput string    `json:"execution_output,omitempty"`
	ExecutionError  string    `json:"execution_error,omitempty"`
}

// --- internal helpers ---

func (s *AgentActionService) getPendingAction(ctx context.Context, orgID uuid.UUID, actionID uuid.UUID) (*models.AgentPendingAction, error) {
	var a models.AgentPendingAction
	err := s.db.QueryRow(ctx, `
		SELECT id, org_id, project_id, feed_card_id, agent_source,
			action_type, action_payload, status, expires_at, created_at
		FROM agent_pending_actions
		WHERE id = $1 AND org_id = $2
	`, actionID, orgID).Scan(
		&a.ID, &a.OrgID, &a.ProjectID, &a.FeedCardID, &a.AgentSource,
		&a.ActionType, &a.ActionPayload, &a.Status, &a.ExpiresAt, &a.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("pending action not found")
		}
		return nil, fmt.Errorf("get pending action: %w", err)
	}
	return &a, nil
}

func (s *AgentActionService) updateStatus(ctx context.Context, actionID uuid.UUID, status models.AgentPendingActionStatus, approvedBy *uuid.UUID, approvedAt *time.Time) error {
	_, err := s.db.Exec(ctx, `
		UPDATE agent_pending_actions
		SET status = $2, approved_by = $3, approved_at = $4
		WHERE id = $1
	`, actionID, status, approvedBy, approvedAt)
	return err
}

func (s *AgentActionService) updateStatusWithReason(ctx context.Context, actionID uuid.UUID, status models.AgentPendingActionStatus, approvedBy *uuid.UUID, approvedAt *time.Time, reason string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE agent_pending_actions
		SET status = $2, approved_by = $3, approved_at = $4, rejection_reason = $5
		WHERE id = $1
	`, actionID, status, approvedBy, approvedAt, reason)
	return err
}
