package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AutopilotPolicy represents an org-level policy for auto-approving agent actions.
type AutopilotPolicy struct {
	ID                  uuid.UUID `json:"id"`
	OrgID               uuid.UUID `json:"org_id"`
	ActionType          string    `json:"action_type"`
	AutoApprove         bool      `json:"auto_approve"`
	MaxCostCents        int64     `json:"max_cost_cents"`
	RequireApprovalFrom []string  `json:"require_approval_from"`
	CooldownMinutes     int       `json:"cooldown_minutes"`
	Enabled             bool      `json:"enabled"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// PolicyDecision is the result of evaluating a policy for an action.
type PolicyDecision struct {
	AutoApproved bool      `json:"auto_approved"`
	Reason       string    `json:"reason"`
	PolicyID     uuid.UUID `json:"policy_id,omitempty"`
}

// PolicyEngine evaluates autopilot policies for agent actions.
type PolicyEngine struct {
	db *pgxpool.Pool
}

// NewPolicyEngine creates a new policy engine.
func NewPolicyEngine(db *pgxpool.Pool) *PolicyEngine {
	return &PolicyEngine{db: db}
}

// Evaluate checks whether an action should be auto-approved based on org policies.
func (e *PolicyEngine) Evaluate(ctx context.Context, orgID uuid.UUID, actionType string, costCents int64) (*PolicyDecision, error) {
	var policy AutopilotPolicy
	err := e.db.QueryRow(ctx, `
		SELECT id, org_id, action_type, auto_approve, max_cost_cents,
			require_approval_from, cooldown_minutes, enabled
		FROM autopilot_policies
		WHERE org_id = $1 AND action_type = $2
	`, orgID, actionType).Scan(
		&policy.ID, &policy.OrgID, &policy.ActionType,
		&policy.AutoApprove, &policy.MaxCostCents,
		&policy.RequireApprovalFrom, &policy.CooldownMinutes, &policy.Enabled,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &PolicyDecision{
				AutoApproved: false,
				Reason:       "no policy configured for this action type",
			}, nil
		}
		return nil, fmt.Errorf("query policy: %w", err)
	}

	if !policy.Enabled {
		return &PolicyDecision{
			AutoApproved: false,
			Reason:       "policy is disabled",
			PolicyID:     policy.ID,
		}, nil
	}

	if !policy.AutoApprove {
		return &PolicyDecision{
			AutoApproved: false,
			Reason:       "policy requires manual approval",
			PolicyID:     policy.ID,
		}, nil
	}

	// Check cost threshold
	if policy.MaxCostCents > 0 && costCents > policy.MaxCostCents {
		return &PolicyDecision{
			AutoApproved: false,
			Reason:       fmt.Sprintf("cost %d exceeds max threshold %d", costCents, policy.MaxCostCents),
			PolicyID:     policy.ID,
		}, nil
	}

	// Check cooldown
	if policy.CooldownMinutes > 0 {
		cooldownMet, err := e.checkCooldown(ctx, orgID, actionType, policy.CooldownMinutes)
		if err != nil {
			slog.Warn("policy cooldown check failed, defaulting to manual", "error", err)
			return &PolicyDecision{
				AutoApproved: false,
				Reason:       "cooldown check failed",
				PolicyID:     policy.ID,
			}, nil
		}
		if !cooldownMet {
			return &PolicyDecision{
				AutoApproved: false,
				Reason:       fmt.Sprintf("cooldown period (%d min) not elapsed", policy.CooldownMinutes),
				PolicyID:     policy.ID,
			}, nil
		}
	}

	return &PolicyDecision{
		AutoApproved: true,
		Reason:       "auto-approved by policy",
		PolicyID:     policy.ID,
	}, nil
}

// checkCooldown verifies that enough time has passed since the last auto-approval.
func (e *PolicyEngine) checkCooldown(ctx context.Context, orgID uuid.UUID, actionType string, cooldownMinutes int) (bool, error) {
	var lastApproval *time.Time
	err := e.db.QueryRow(ctx, `
		SELECT MAX(created_at)
		FROM agent_pending_actions
		WHERE org_id = $1 AND action_type = $2 AND status = 'approved'
		  AND created_at > NOW() - INTERVAL '1 minute' * $3
	`, orgID, actionType, cooldownMinutes).Scan(&lastApproval)
	if err != nil {
		return false, err
	}
	return lastApproval == nil, nil
}

// ListPolicies returns all policies for an org.
func (e *PolicyEngine) ListPolicies(ctx context.Context, orgID uuid.UUID) ([]AutopilotPolicy, error) {
	rows, err := e.db.Query(ctx, `
		SELECT id, org_id, action_type, auto_approve, max_cost_cents,
			require_approval_from, cooldown_minutes, enabled, created_at, updated_at
		FROM autopilot_policies
		WHERE org_id = $1
		ORDER BY action_type
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []AutopilotPolicy
	for rows.Next() {
		var p AutopilotPolicy
		if err := rows.Scan(&p.ID, &p.OrgID, &p.ActionType, &p.AutoApprove,
			&p.MaxCostCents, &p.RequireApprovalFrom, &p.CooldownMinutes,
			&p.Enabled, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, nil
}

// UpsertPolicy creates or updates a policy for an action type.
func (e *PolicyEngine) UpsertPolicy(ctx context.Context, orgID uuid.UUID, actionType string, autoApprove bool, maxCostCents int64, cooldownMinutes int) (*AutopilotPolicy, error) {
	var p AutopilotPolicy
	err := e.db.QueryRow(ctx, `
		INSERT INTO autopilot_policies (org_id, action_type, auto_approve, max_cost_cents, cooldown_minutes)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (org_id, action_type) DO UPDATE SET
			auto_approve = EXCLUDED.auto_approve,
			max_cost_cents = EXCLUDED.max_cost_cents,
			cooldown_minutes = EXCLUDED.cooldown_minutes,
			updated_at = NOW()
		RETURNING id, org_id, action_type, auto_approve, max_cost_cents,
			require_approval_from, cooldown_minutes, enabled, created_at, updated_at
	`, orgID, actionType, autoApprove, maxCostCents, cooldownMinutes).Scan(
		&p.ID, &p.OrgID, &p.ActionType, &p.AutoApprove,
		&p.MaxCostCents, &p.RequireApprovalFrom, &p.CooldownMinutes,
		&p.Enabled, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert policy: %w", err)
	}
	return &p, nil
}
