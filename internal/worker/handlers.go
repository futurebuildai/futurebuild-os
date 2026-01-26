package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/colton/futurebuild/internal/agents"
	"github.com/colton/futurebuild/internal/futureshade"
	"github.com/colton/futurebuild/internal/futureshade/gateway"
	"github.com/colton/futurebuild/internal/futureshade/skills"
	"github.com/colton/futurebuild/internal/futureshade/tribunal"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/clock"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkerHandler struct {
	focusAgent       *agents.DailyFocusAgent
	procurementAgent *agents.ProcurementAgent
	db               *pgxpool.Pool
	clock            clock.Clock
	// FutureShade Action Bridge fields (optional - initialized via WithSkillExecution)
	skillRegistry     *skills.Registry
	executionRepo     *gateway.Repository
	futureShadeConfig futureshade.Config
	// Automated PR Review fields (optional - initialized via WithPRReview)
	githubService service.GitHubServicer
	tribunalEngine *tribunal.ConsensusEngine
	tribunalRepo   *tribunal.Repository
}

func NewWorkerHandler(focusAgent *agents.DailyFocusAgent, procurementAgent *agents.ProcurementAgent, db *pgxpool.Pool, clk clock.Clock) *WorkerHandler {
	return &WorkerHandler{
		focusAgent:       focusAgent,
		procurementAgent: procurementAgent,
		db:               db,
		clock:            clk,
	}
}

// WithSkillExecution configures the handler for FutureShade skill execution.
// See specs/FUTURESHADE_AGENTS_SPEC.md Section 4 (Action Bridge)
func (h *WorkerHandler) WithSkillExecution(registry *skills.Registry, repo *gateway.Repository, config futureshade.Config) *WorkerHandler {
	h.skillRegistry = registry
	h.executionRepo = repo
	h.futureShadeConfig = config
	return h
}

// WithPRReview configures the handler for Automated PR Review.
// See docs/AUTOMATED_PR_REVIEW_PRD.md
func (h *WorkerHandler) WithPRReview(githubService service.GitHubServicer, engine *tribunal.ConsensusEngine, repo *tribunal.Repository) *WorkerHandler {
	h.githubService = githubService
	h.tribunalEngine = engine
	h.tribunalRepo = repo
	return h
}

// HandleDailyBriefing executes the daily briefing agent logic.
func (h *WorkerHandler) HandleDailyBriefing(ctx context.Context, task *asynq.Task) error {
	slog.Info("Handling Daily Briefing Task...")
	if err := h.focusAgent.Execute(ctx); err != nil {
		return fmt.Errorf("daily briefing agent failed: %w", err)
	}
	slog.Info("Daily Briefing Task Completed Successfully.")
	return nil
}

// HandleProcurementCheck executes the procurement agent logic.
// See PRODUCTION_PLAN.md Step 46
func (h *WorkerHandler) HandleProcurementCheck(ctx context.Context, task *asynq.Task) error {
	slog.Info("Handling Procurement Check Task...")
	if err := h.procurementAgent.Execute(ctx); err != nil {
		return fmt.Errorf("procurement agent failed: %w", err)
	}
	slog.Info("Procurement Check Task Completed Successfully.")
	return nil
}

// HandleHydrateProject handles project-scoped hydration of procurement items.
// Triggered when a new project is created (event-driven).
// P1 Performance Fix: Replaces inefficient cron-swept hydration.
func (h *WorkerHandler) HandleHydrateProject(ctx context.Context, task *asynq.Task) error {
	var payload HydrateProjectPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("invalid hydrate project payload: %w", err)
	}

	logArgs := []any{"project_id", payload.ProjectID}
	slog.Info("Handling Hydrate Project Task", logArgs...)
	if err := h.procurementAgent.HydrateProject(ctx, payload.ProjectID); err != nil {
		return fmt.Errorf("hydrate project failed: %w", err)
	}
	slog.Info("Hydrate Project Task Completed Successfully", logArgs...)
	return nil
}

// HandleProcurementNotification processes async procurement notifications.
// P1 Performance Fix: Sidecar pattern - notifications are queued instead of written synchronously.
// This handler performs dampening check and logs to communication_logs.
func (h *WorkerHandler) HandleProcurementNotification(ctx context.Context, task *asynq.Task) error {
	var payload ProcurementNotificationPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("invalid notification payload: %w", err)
	}

	slog.Info("Handling Procurement Notification", "item_id", payload.ItemID)

	// Dampening check: Skip if notification was sent in last 72 hours
	shouldSend, err := h.shouldSendNotification(ctx, payload.ItemID, payload.Timestamp)
	if err != nil {
		return fmt.Errorf("dampening check failed: %w", err)
	}
	if !shouldSend {
		slog.Info("Notification dampened (already sent within 72h)", "item_id", payload.ItemID)
		return nil
	}

	// Log notification to communication_logs
	if err := h.logNotification(ctx, payload); err != nil {
		return fmt.Errorf("log notification failed: %w", err)
	}

	slog.Info("Procurement Notification sent", "item_id", payload.ItemID)
	return nil
}

// shouldSendNotification checks communication_logs for recent alerts.
// See User Amendment #4: 72-hour dampening
func (h *WorkerHandler) shouldSendNotification(ctx context.Context, itemID interface{}, timestamp interface{}) (bool, error) {
	query := `
		SELECT COUNT(*) FROM communication_logs
		WHERE related_entity_id = $1
		  AND timestamp > $2 - INTERVAL '72 hours'
		  AND direction = 'Outbound'
	`
	var count int
	err := h.db.QueryRow(ctx, query, itemID, timestamp).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// logNotification persists the alert to communication_logs.
func (h *WorkerHandler) logNotification(ctx context.Context, payload ProcurementNotificationPayload) error {
	query := `
		INSERT INTO communication_logs (
			project_id, direction, content, channel, timestamp,
			related_entity_id, related_entity_type
		)
		SELECT p.id, 'Outbound', $1, 'Chat', $2, $3, 'procurement_item'
		FROM procurement_items pi
		JOIN project_tasks pt ON pi.project_task_id = pt.id
		JOIN projects p ON pt.project_id = p.id
		WHERE pi.id = $4
	`
	content := fmt.Sprintf("[PROCUREMENT ALERT] %s", payload.Message)
	_, err := h.db.Exec(ctx, query, content, payload.Timestamp, payload.ItemID, payload.ItemID)
	return err
}

// HandleSkillExecution processes FutureShade skill execution tasks.
// Implements idempotency (skips if not PENDING) and circuit breaker (skips if disabled).
// See specs/FUTURESHADE_AGENTS_SPEC.md Section 4 (Action Bridge)
func (h *WorkerHandler) HandleSkillExecution(ctx context.Context, task *asynq.Task) error {
	// Parse payload
	var payload SkillExecutionPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("invalid skill execution payload: %w", err)
	}

	logArgs := []any{
		"execution_id", payload.ExecutionID,
		"decision_id", payload.DecisionID,
		"skill_id", payload.SkillID,
	}

	// Circuit Breaker: Check if FutureShade is enabled
	if !h.futureShadeConfig.Enabled {
		slog.Info("FutureShade disabled, skipping skill execution", logArgs...)
		return nil
	}

	// Validate handler was configured
	if h.skillRegistry == nil || h.executionRepo == nil {
		slog.Error("skill execution handler not configured", logArgs...)
		return fmt.Errorf("skill execution handler not configured")
	}

	// Idempotency Check: Skip if not PENDING
	status, err := h.executionRepo.GetStatus(ctx, payload.ExecutionID)
	if err != nil {
		slog.Error("failed to get execution status", append(logArgs, "error", err)...)
		return fmt.Errorf("get execution status: %w", err)
	}
	if status != gateway.StatusPending {
		slog.Info("skill execution already processed, skipping",
			append(logArgs, "current_status", status)...)
		return nil
	}

	// Mark as RUNNING
	if err := h.executionRepo.MarkRunning(ctx, payload.ExecutionID); err != nil {
		slog.Error("failed to mark execution as running", append(logArgs, "error", err)...)
		return fmt.Errorf("mark running: %w", err)
	}

	slog.Info("Starting skill execution", logArgs...)

	// Get skill from registry
	skill, ok := h.skillRegistry.Get(payload.SkillID)
	if !ok {
		errMsg := fmt.Sprintf("skill %q not found in registry", payload.SkillID)
		if err := h.executionRepo.UpdateExecutionStatus(ctx, payload.ExecutionID,
			gateway.StatusFailed, nil, &errMsg); err != nil {
			slog.Error("failed to update execution status", append(logArgs, "error", err)...)
		}
		return fmt.Errorf("skill %q not found in registry", payload.SkillID)
	}

	// Execute the skill
	result, execErr := skill.Execute(ctx, payload.Params)

	// Update execution status based on result
	if execErr != nil {
		errMsg := execErr.Error()
		if err := h.executionRepo.UpdateExecutionStatus(ctx, payload.ExecutionID,
			gateway.StatusFailed, &result.Summary, &errMsg); err != nil {
			slog.Error("failed to update execution status on failure",
				append(logArgs, "error", err)...)
		}
		slog.Error("Skill execution failed", append(logArgs, "error", execErr)...)
		return fmt.Errorf("skill execution failed: %w", execErr)
	}

	// Success
	if err := h.executionRepo.UpdateExecutionStatus(ctx, payload.ExecutionID,
		gateway.StatusCompleted, &result.Summary, nil); err != nil {
		slog.Error("failed to update execution status on success",
			append(logArgs, "error", err)...)
	}

	slog.Info("Skill execution completed successfully",
		append(logArgs, "summary", result.Summary)...)

	return nil
}

// HandleReviewPR processes Automated PR Review tasks.
// Fetches PR diff, consults Tribunal, and posts verdict as PR comment.
// See docs/AUTOMATED_PR_REVIEW_PRD.md
func (h *WorkerHandler) HandleReviewPR(ctx context.Context, task *asynq.Task) error {
	var payload ReviewPRPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("invalid review_pr payload: %w", err)
	}

	logArgs := []any{
		"case_id", payload.CaseID,
		"owner", payload.Owner,
		"repo", payload.Repo,
		"pr_number", payload.PRNumber,
	}

	// Validate handler configuration
	if h.githubService == nil || h.tribunalEngine == nil || h.tribunalRepo == nil {
		slog.Error("worker/review_pr: handler not configured", logArgs...)
		return fmt.Errorf("PR review handler not configured")
	}

	// Step 1: Idempotency check - skip if already processed
	exists, err := h.tribunalRepo.DecisionExistsByCaseID(ctx, payload.CaseID)
	if err != nil {
		slog.Error("worker/review_pr: idempotency check failed", append(logArgs, "error", err)...)
		return fmt.Errorf("idempotency check failed: %w", err)
	}
	if exists {
		slog.Info("worker/review_pr: skipping duplicate", logArgs...)
		return nil
	}

	slog.Info("worker/review_pr: processing PR", logArgs...)

	// Step 2: Fetch PR diff
	diff, err := h.githubService.FetchPRDiff(ctx, payload.Owner, payload.Repo, payload.PRNumber)
	if err != nil {
		slog.Error("worker/review_pr: fetch diff failed", append(logArgs, "error", err)...)
		return fmt.Errorf("fetch PR diff: %w", err)
	}

	// Step 3: Sanitize diff (additional layer - service already sanitizes)
	sanitizedDiff := sanitizePRDiff(diff)

	// Step 4: Build Tribunal request
	req := tribunal.TribunalRequest{
		CaseID:  payload.CaseID,
		Intent:  fmt.Sprintf("Automated security and consistency audit for PR #%d: %s", payload.PRNumber, payload.PRTitle),
		Context: sanitizedDiff,
	}

	// Step 5: Consult Tribunal
	resp, err := h.tribunalEngine.Review(ctx, req)
	if err != nil {
		slog.Error("worker/review_pr: tribunal review failed", append(logArgs, "error", err)...)
		return fmt.Errorf("tribunal review: %w", err)
	}

	slog.Info("worker/review_pr: tribunal decision",
		append(logArgs,
			"decision_id", resp.DecisionID,
			"status", resp.Status,
			"consensus_score", resp.ConsensusScore,
		)...)

	// Step 6: Format and post PR comment
	comment := formatPRComment(resp)
	if err := h.githubService.PostPRComment(ctx, payload.Owner, payload.Repo, payload.PRNumber, comment); err != nil {
		slog.Error("worker/review_pr: post comment failed", append(logArgs, "error", err)...)
		return fmt.Errorf("post PR comment: %w", err)
	}

	slog.Info("worker/review_pr: completed successfully", logArgs...)
	return nil
}

// sanitizePRDiff performs additional sanitization on PR diff content.
// Removes potential prompt injection patterns.
func sanitizePRDiff(diff string) string {
	// Remove standalone delimiter lines that could confuse AI models
	lines := strings.Split(diff, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" || trimmed == "===" || trimmed == ">>>" {
			lines[i] = "[separator]"
		}
	}
	return strings.Join(lines, "\n")
}

// formatPRComment formats the Tribunal response as a GitHub PR comment.
// See docs/AUTOMATED_PR_REVIEW_PRD.md "PR Comment Format"
func formatPRComment(resp *tribunal.TribunalResponse) string {
	var statusEmoji string
	switch resp.Status {
	case tribunal.DecisionApproved:
		statusEmoji = ":white_check_mark:"
	case tribunal.DecisionRejected:
		statusEmoji = ":x:"
	case tribunal.DecisionConflict:
		statusEmoji = ":warning:"
	default:
		statusEmoji = ":question:"
	}

	return fmt.Sprintf(`## FutureBuild AI Review %s

**Status**: %s
**Consensus Score**: %.2f

### Summary
%s

### Recommendations
%s

---
*Generated by [The Tribunal](https://github.com/colton/futurebuild) | Decision ID: %s*`,
		statusEmoji,
		resp.Status,
		resp.ConsensusScore,
		resp.Summary,
		resp.Plan,
		resp.DecisionID.String(),
	)
}
