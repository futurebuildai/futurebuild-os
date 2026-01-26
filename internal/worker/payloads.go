package worker

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// Task Types
// See BACKEND_SCOPE.md Section 3.5 (Action Engine)
const (
	TypeDailyBriefing           = "task:daily_briefing"
	TypeProcurementCheck        = "task:procurement_check"
	TypeHydrateProject          = "task:hydrate_project"
	TypeProcurementNotification = "task:procurement_notification"
	TypeSkillExecution          = "task:skill_execution" // FutureShade Action Bridge
	TypeReviewPR                = "task:review_pr"       // Automated PR Review
)

// HydrateProjectPayload contains the project ID for scoped hydration.
// See implementation_plan.md: "Event-Driven Hydration"
type HydrateProjectPayload struct {
	ProjectID uuid.UUID `json:"project_id"`
}

// NewDailyBriefingTask creates a task for daily briefing.
// No payload needed for a global daily briefing trigger.
func NewDailyBriefingTask() *asynq.Task {
	return asynq.NewTask(TypeDailyBriefing, nil)
}

// NewProcurementCheckTask creates a task for procurement analysis.
// See PRODUCTION_PLAN.md Step 46
func NewProcurementCheckTask() *asynq.Task {
	return asynq.NewTask(TypeProcurementCheck, nil)
}

// NewHydrateProjectTask creates a task for project-scoped hydration.
// Enqueued when a project is created to populate procurement_items.
// P1 Performance Fix: Replaces cron-swept hydration.
func NewHydrateProjectTask(projectID uuid.UUID) (*asynq.Task, error) {
	payload, err := json.Marshal(HydrateProjectPayload{ProjectID: projectID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeHydrateProject, payload), nil
}

// ProcurementNotificationPayload contains notification data for async delivery.
// P1 Performance Fix: Enables sidecar notification pattern to reduce DB round-trips.
type ProcurementNotificationPayload struct {
	ItemID    uuid.UUID `json:"item_id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// NewProcurementNotificationTask creates a task for async notification delivery.
// P1 Performance Fix: Notifications are queued instead of written synchronously.
func NewProcurementNotificationTask(itemID uuid.UUID, message string, ts time.Time) (*asynq.Task, error) {
	payload, err := json.Marshal(ProcurementNotificationPayload{
		ItemID:    itemID,
		Message:   message,
		Timestamp: ts,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeProcurementNotification, payload, asynq.Queue("default")), nil
}

// SkillExecutionPayload contains data for FutureShade skill execution tasks.
// See specs/FUTURESHADE_AGENTS_SPEC.md Section 4 (Action Bridge)
type SkillExecutionPayload struct {
	DecisionID  uuid.UUID              `json:"decision_id"`
	ExecutionID uuid.UUID              `json:"execution_id"`
	SkillID     string                 `json:"skill_id"`
	Params      map[string]interface{} `json:"params"`
}

// NewSkillExecutionTask creates a task for FutureShade skill execution.
// Used by ExecutionGateway to enqueue Tribunal-triggered actions.
func NewSkillExecutionTask(decisionID, executionID uuid.UUID, skillID string, params map[string]interface{}) (*asynq.Task, error) {
	payload, err := json.Marshal(SkillExecutionPayload{
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

// ReviewPRPayload contains data for GitHub PR review tasks.
// See docs/AUTOMATED_PR_REVIEW_PRD.md
type ReviewPRPayload struct {
	CaseID   string `json:"case_id"`    // Format: GH_{owner}/{repo}#{number}_{sha}
	Owner    string `json:"owner"`      // Repository owner
	Repo     string `json:"repo"`       // Repository name
	PRNumber int    `json:"pr_number"`  // Pull request number
	HeadSHA  string `json:"head_sha"`   // HEAD commit SHA
	PRTitle  string `json:"pr_title"`   // Pull request title
}

// NewReviewPRTask creates a task for Automated PR Review.
// Enqueued by GitHub webhook handler when a PR is opened/synchronized/reopened.
func NewReviewPRTask(payload ReviewPRPayload) (*asynq.Task, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeReviewPR, payloadBytes, asynq.Queue("default")), nil
}
