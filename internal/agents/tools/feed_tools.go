package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/google/uuid"
)

// FeedCardWriter allows tools to write cards into the portfolio feed.
// Mirrors agents.FeedWriter to avoid import cycles. Go's implicit interfaces
// mean *service.FeedService satisfies both.
type FeedCardWriter interface {
	WriteCard(ctx context.Context, card *models.FeedCard) error
}

// PendingActionCreator allows tools to create pending action records.
// Satisfied by *service.AgentActionService.
type PendingActionCreator interface {
	CreatePendingAction(ctx context.Context, action *models.AgentPendingAction) error
}

// PolicyEvaluator evaluates autopilot policies for auto-approval.
// Satisfied by *service.PolicyEngine.
type PolicyEvaluator interface {
	Evaluate(ctx context.Context, orgID uuid.UUID, actionType string, costCents int64) (*PolicyDecision, error)
}

// PolicyDecision mirrors service.PolicyDecision to avoid import cycles.
type PolicyDecision struct {
	AutoApproved bool      `json:"auto_approved"`
	Reason       string    `json:"reason"`
	PolicyID     uuid.UUID `json:"policy_id,omitempty"`
}

// ActionExecutor executes an approved action by tool name and payload.
// Satisfied by tools.ActionRunnerAdapter.
type ActionExecutor interface {
	Execute(ctx context.Context, toolName string, payload json.RawMessage) (string, error)
}

// FeedToolsConfig holds optional dependencies for feed tools.
type FeedToolsConfig struct {
	ActionCreator  PendingActionCreator
	PolicyEngine   PolicyEvaluator
	ActionExecutor ActionExecutor
}

// RegisterFeedTools registers feed card and approval-related tools.
func RegisterFeedTools(r *Registry, feedWriter FeedCardWriter, pendingActions ...PendingActionCreator) {
	var actionCreator PendingActionCreator
	if len(pendingActions) > 0 {
		actionCreator = pendingActions[0]
	}
	RegisterFeedToolsWithPolicy(r, feedWriter, FeedToolsConfig{ActionCreator: actionCreator})
}

// RegisterFeedToolsWithPolicy registers feed tools with optional autopilot policy support.
func RegisterFeedToolsWithPolicy(r *Registry, feedWriter FeedCardWriter, cfg FeedToolsConfig) {
	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "write_feed_card",
			Description: "Create an informational feed card that appears in the user's portfolio feed. Use for status updates, notifications, and non-actionable information.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"card_type":{"type":"string","description":"Type of card (e.g., daily_briefing, weather_risk, milestone)"},"headline":{"type":"string","description":"Short headline (max 120 chars)"},"body":{"type":"string","description":"Card body with details"},"priority":{"type":"integer","enum":[0,1,2,3],"description":"0=critical, 1=urgent, 2=normal, 3=low"},"horizon":{"type":"string","enum":["today","this_week","horizon"],"description":"When this is relevant"}},"required":["card_type","headline","body","priority","horizon"]}`),
		},
		Handler: func(ctx context.Context, input json.RawMessage) (string, error) {
			scope := MustGetScope(ctx)
			var params struct {
				CardType string `json:"card_type"`
				Headline string `json:"headline"`
				Body     string `json:"body"`
				Priority int    `json:"priority"`
				Horizon  string `json:"horizon"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return "", fmt.Errorf("parse input: %w", err)
			}

			agentSource := "ChatAgent"
			card := &models.FeedCard{
				OrgID:       scope.OrgID,
				ProjectID:   scope.ProjectID,
				CardType:    models.FeedCardType(params.CardType),
				Priority:    params.Priority,
				Headline:    params.Headline,
				Body:        params.Body,
				Horizon:     models.FeedCardHorizon(params.Horizon),
				AgentSource: &agentSource,
			}
			if err := feedWriter.WriteCard(ctx, card); err != nil {
				return "", fmt.Errorf("write feed card: %w", err)
			}
			return fmt.Sprintf(`{"success":true,"card_id":"%s"}`, card.ID), nil
		},
	})

	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name: "create_approval_card",
			Description: `Create an approval card that requires human confirmation before an action is executed. Use this for ANY state-changing operation: updating task status, sending notifications, making schedule changes, approving change orders, etc.

The card appears in the user's feed with Approve/Reject buttons. When approved, the stored action is executed automatically.

Examples:
- "Recommend delaying framing start by 2 days due to rain" → action_type: "update_task_duration", action_payload: {...}
- "Send confirmation request to electrician" → action_type: "send_sms", action_payload: {...}
- "Order custom windows now (lead time closing)" → action_type: "mark_ordered", action_payload: {...}`,
			InputSchema: json.RawMessage(`{"type":"object","properties":{"headline":{"type":"string","description":"Clear description of what action is being recommended"},"body":{"type":"string","description":"Detailed reasoning: why this action is recommended, what happens if ignored"},"consequence":{"type":"string","description":"What happens if this action is NOT taken (e.g., '2-day critical path slip')"},"priority":{"type":"integer","enum":[0,1,2,3],"description":"0=critical, 1=urgent, 2=normal, 3=low"},"action_type":{"type":"string","description":"The tool/action to execute on approval (e.g., 'update_task_status', 'send_email', 'send_sms', 'recalculate_schedule')"},"action_payload":{"type":"object","description":"The exact parameters to pass to the action tool when approved"},"expires_hours":{"type":"integer","description":"Hours until this approval expires (default: 24)","default":24}},"required":["headline","body","priority","action_type","action_payload"]}`),
		},
		Handler: func(ctx context.Context, input json.RawMessage) (string, error) {
			scope := MustGetScope(ctx)
			var params struct {
				Headline      string          `json:"headline"`
				Body          string          `json:"body"`
				Consequence   string          `json:"consequence"`
				Priority      int             `json:"priority"`
				ActionType    string          `json:"action_type"`
				ActionPayload json.RawMessage `json:"action_payload"`
				ExpiresHours  int             `json:"expires_hours"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return "", fmt.Errorf("parse input: %w", err)
			}

			if params.ExpiresHours <= 0 {
				params.ExpiresHours = 24
			}
			expiresAt := time.Now().UTC().Add(time.Duration(params.ExpiresHours) * time.Hour)

			agentSource := "ChatAgent"
			var consequence *string
			if params.Consequence != "" {
				consequence = &params.Consequence
			}

			// Store action_type + action_payload in engine_data for the feed card
			engineData, _ := json.Marshal(map[string]interface{}{
				"action_type":    params.ActionType,
				"action_payload": params.ActionPayload,
			})

			card := &models.FeedCard{
				ID:          uuid.New(),
				OrgID:       scope.OrgID,
				ProjectID:   scope.ProjectID,
				CardType:    models.FeedCardAgentApproval,
				Priority:    params.Priority,
				Headline:    params.Headline,
				Body:        params.Body,
				Consequence: consequence,
				Horizon:     models.FeedCardHorizonToday,
				AgentSource: &agentSource,
				ExpiresAt:   &expiresAt,
				EngineData:  engineData,
				Actions: []models.FeedCardAction{
					{ID: "approve_agent_action", Label: "Approve", Style: "primary"},
					{ID: "reject_agent_action", Label: "Reject", Style: "danger"},
					{ID: "modify", Label: "Modify", Style: "secondary"},
				},
			}

			// Check autopilot policy — if auto-approved, execute immediately
			if cfg.PolicyEngine != nil && cfg.ActionExecutor != nil {
				decision, pErr := cfg.PolicyEngine.Evaluate(ctx, scope.OrgID, params.ActionType, 0)
				if pErr == nil && decision.AutoApproved {
					// Auto-execute the action
					result, execErr := cfg.ActionExecutor.Execute(ctx, params.ActionType, params.ActionPayload)
					if execErr != nil {
						slog.Warn("autopilot: auto-execution failed, falling back to manual approval",
							"action_type", params.ActionType, "error", execErr)
					} else {
						slog.Info("autopilot: action auto-approved and executed",
							"action_type", params.ActionType, "policy_id", decision.PolicyID)
						return fmt.Sprintf(`{"success":true,"auto_approved":true,"reason":"%s","result":%s}`,
							decision.Reason, result), nil
					}
				}
			}

			if err := feedWriter.WriteCard(ctx, card); err != nil {
				return "", fmt.Errorf("write approval card: %w", err)
			}

			// Create the pending action record (human-in-the-loop)
			if cfg.ActionCreator != nil {
				pendingAction := &models.AgentPendingAction{
					OrgID:         scope.OrgID,
					ProjectID:     scope.ProjectID,
					FeedCardID:    card.ID,
					AgentSource:   agentSource,
					ActionType:    params.ActionType,
					ActionPayload: params.ActionPayload,
					Status:        models.PendingActionStatusPending,
					ExpiresAt:     &expiresAt,
				}
				if err := cfg.ActionCreator.CreatePendingAction(ctx, pendingAction); err != nil {
					slog.Warn("failed to create pending action record", "error", err)
				}
			}

			return fmt.Sprintf(`{"success":true,"approval_card_id":"%s","action_type":"%s","expires_at":"%s"}`,
				card.ID, params.ActionType, expiresAt.Format(time.RFC3339)), nil
		},
	})
}

// RegisterInvoiceTools registers invoice-related tools.
// Placeholder for Phase 6 enhancement.
func RegisterInvoiceTools(r *Registry) {
	// Invoice tools will be registered when the InvoiceServicer is available
}
