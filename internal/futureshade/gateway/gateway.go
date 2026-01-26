package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/colton/futurebuild/internal/futureshade"
	"github.com/colton/futurebuild/internal/futureshade/skills"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// PlanAction represents a single action in a remediation plan.
type PlanAction struct {
	SkillID string         `json:"skill_id"`
	Params  map[string]any `json:"params"`
}

// RemediationPlan represents a Tribunal-generated plan of actions.
type RemediationPlan struct {
	Actions []PlanAction `json:"actions"`
}

// ExecutionGateway connects Tribunal decisions to skill execution via Asynq workers.
// See specs/FUTURESHADE_AGENTS_SPEC.md Section 4 (Action Bridge)
type ExecutionGateway struct {
	config   futureshade.Config
	registry *skills.Registry
	repo     *Repository
	client   *asynq.Client
}

// NewExecutionGateway creates a new execution gateway.
func NewExecutionGateway(
	config futureshade.Config,
	registry *skills.Registry,
	repo *Repository,
	redisAddr string,
) *ExecutionGateway {
	return &ExecutionGateway{
		config:   config,
		registry: registry,
		repo:     repo,
		client:   asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr}),
	}
}

// Close releases the underlying Asynq client resources.
func (g *ExecutionGateway) Close() error {
	return g.client.Close()
}

// EnqueuePlan parses a remediation plan JSON and enqueues skill execution tasks.
// Returns an error if any skill_id in the plan is not registered (validation-first).
// All actions are atomically validated before any are enqueued.
//
// Circuit Breaker: Returns nil immediately if FutureShade is disabled.
func (g *ExecutionGateway) EnqueuePlan(ctx context.Context, decisionID uuid.UUID, planJSON []byte) error {
	// Circuit Breaker: Check if FutureShade is enabled
	if !g.config.Enabled {
		slog.Info("FutureShade disabled, skipping plan enqueue",
			"decision_id", decisionID,
		)
		return nil
	}

	// Parse the remediation plan
	var plan RemediationPlan
	if err := json.Unmarshal(planJSON, &plan); err != nil {
		return fmt.Errorf("parse remediation plan: %w", err)
	}

	if len(plan.Actions) == 0 {
		slog.Debug("empty remediation plan, nothing to enqueue",
			"decision_id", decisionID,
		)
		return nil
	}

	// Validation-First: Check all skill_ids exist before enqueuing any
	var unknownSkills []string
	for _, action := range plan.Actions {
		if !g.registry.Has(action.SkillID) {
			unknownSkills = append(unknownSkills, action.SkillID)
		}
	}
	if len(unknownSkills) > 0 {
		return fmt.Errorf("unknown skill IDs in plan: %v", unknownSkills)
	}

	// Create execution logs and enqueue tasks
	var enqueued int
	for _, action := range plan.Actions {
		// Create PENDING execution log
		execID, err := g.repo.CreateExecutionLog(ctx, decisionID, action.SkillID, action.Params)
		if err != nil {
			slog.Error("failed to create execution log",
				"decision_id", decisionID,
				"skill_id", action.SkillID,
				"error", err,
			)
			// Continue with other actions - partial success is acceptable
			continue
		}

		// Create and enqueue Asynq task
		task, err := NewSkillExecutionTask(decisionID, execID, action.SkillID, action.Params)
		if err != nil {
			slog.Error("failed to create skill execution task",
				"execution_id", execID,
				"skill_id", action.SkillID,
				"error", err,
			)
			continue
		}

		_, err = g.client.EnqueueContext(ctx, task, asynq.Queue("default"))
		if err != nil {
			slog.Error("failed to enqueue skill execution task",
				"execution_id", execID,
				"skill_id", action.SkillID,
				"error", err,
			)
			continue
		}

		enqueued++
		slog.Debug("enqueued skill execution",
			"decision_id", decisionID,
			"execution_id", execID,
			"skill_id", action.SkillID,
		)
	}

	slog.Info("remediation plan enqueued",
		"decision_id", decisionID,
		"total_actions", len(plan.Actions),
		"enqueued", enqueued,
	)

	return nil
}

// EnqueuePlanFromJSON is a convenience method that takes a JSON string.
func (g *ExecutionGateway) EnqueuePlanFromJSON(ctx context.Context, decisionID uuid.UUID, planJSON string) error {
	return g.EnqueuePlan(ctx, decisionID, []byte(planJSON))
}

// ValidatePlan validates a remediation plan without enqueuing it.
// Useful for pre-flight validation before Tribunal approval.
func (g *ExecutionGateway) ValidatePlan(planJSON []byte) error {
	var plan RemediationPlan
	if err := json.Unmarshal(planJSON, &plan); err != nil {
		return fmt.Errorf("parse remediation plan: %w", err)
	}

	var unknownSkills []string
	for _, action := range plan.Actions {
		if !g.registry.Has(action.SkillID) {
			unknownSkills = append(unknownSkills, action.SkillID)
		}
	}
	if len(unknownSkills) > 0 {
		return errors.New("unknown skill IDs: " + fmt.Sprintf("%v", unknownSkills))
	}

	return nil
}

// TypeSkillExecution is the Asynq task type for skill execution.
// Must match worker.TypeSkillExecution.
const TypeSkillExecution = "task:skill_execution"

// skillExecutionPayload is the internal payload structure.
// Must match worker.SkillExecutionPayload.
type skillExecutionPayload struct {
	DecisionID  uuid.UUID      `json:"decision_id"`
	ExecutionID uuid.UUID      `json:"execution_id"`
	SkillID     string         `json:"skill_id"`
	Params      map[string]any `json:"params"`
}

// NewSkillExecutionTask creates an Asynq task for skill execution.
// Package-local to avoid circular dependency with worker package.
func NewSkillExecutionTask(decisionID, executionID uuid.UUID, skillID string, params map[string]any) (*asynq.Task, error) {
	payload, err := json.Marshal(skillExecutionPayload{
		DecisionID:  decisionID,
		ExecutionID: executionID,
		SkillID:     skillID,
		Params:      params,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSkillExecution, payload, asynq.Queue("default")), nil
}
