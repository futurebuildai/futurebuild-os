package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
)

// subLiaisonClaudeSystemPrompt guides Claude for subcontractor message understanding.
const subLiaisonClaudeSystemPrompt = `You are a construction superintendent AI that understands subcontractor communications.

Your job: Parse inbound SMS/email messages from subcontractors and determine the correct action.

## Message Types You'll See
Subcontractors reply to automated confirmation requests with natural language. Examples:
- "Yeah we'll be there Monday" → confirmation
- "Running about 2 hours behind" → delay with estimate
- "We're at 75%, should wrap up tomorrow" → progress update (75%)
- "Can't make it, truck broke down" → delay, needs escalation
- "We're done, ready for inspection" → progress update (100%)
- "Who do I talk to about the change order?" → question, needs escalation

## Your Task
For each inbound message, call parse_inbound_message with structured data extracted from the text.

Then, based on the situation:
1. For simple confirmations or progress updates: call create_approval_card with a low-priority informational card
2. For delays with schedule impact: call create_approval_card with an urgent card recommending follow-up action
3. For cancellations or serious issues: call create_approval_card with a critical card recommending immediate PM notification
4. For ambiguous messages or questions: call create_approval_card recommending the PM review the message

## Be construction-aware:
- "Running behind" usually means 1-4 hours, not days
- "Can't make it" for a scheduled task is more serious than a delay
- Weather-related delays are common and usually short-term
- Subcontractors often text informally — interpret generously
`

// parsedInbound represents Claude's structured understanding of an inbound message.
type parsedInbound struct {
	ProgressPercent *int    `json:"progress_percent,omitempty"`
	IsConfirmation  bool    `json:"is_confirmation"`
	IsDelay         bool    `json:"is_delay"`
	IsCancellation  bool    `json:"is_cancellation"`
	IsQuestion      bool    `json:"is_question"`
	EstimatedDelay  string  `json:"estimated_delay,omitempty"`
	Summary         string  `json:"summary"`
	Urgency         string  `json:"urgency"` // "low", "medium", "high", "critical"
	SuggestedAction string  `json:"suggested_action,omitempty"`
}

// handleInboundWithClaude uses Claude to understand a nuanced inbound message and take action.
// Falls back to regex-based parsing on failure.
func (a *SubLiaisonAgent) handleInboundWithClaude(
	ctx context.Context,
	contact *types.Contact,
	taskID *uuid.UUID,
	task *taskDetails,
	body string,
) {
	userMessage := fmt.Sprintf(`Parse this inbound message from subcontractor %s (%s) regarding task "%s":

Message: "%s"

Determine the intent and create appropriate feed cards.`, contact.Name, contact.Company, task.Name, body)

	projectCtx := ProjectContext{
		ProjectID: task.ProjectID,
		OrgID:     task.OrgID,
		UserID:    uuid.Nil, // Agent user
	}

	result, err := a.claudeRunner.Run(ctx, subLiaisonClaudeSystemPrompt, userMessage, projectCtx)
	if err != nil {
		slog.Warn("claude sub liaison reasoning failed, falling back to regex parsing",
			"task_id", taskID, "contact_id", contact.ID, "error", err)
		a.handleInboundFallback(ctx, contact, taskID, task, body)
		return
	}

	slog.Info("claude sub liaison reasoning completed",
		"task_id", taskID,
		"contact", contact.Name,
		"turns", result.Turns,
		"tools_used", result.ToolsUsed,
		"tokens", result.TotalTokens,
	)
}

// handleInboundFallback uses the original regex-based parsing when Claude is unavailable.
func (a *SubLiaisonAgent) handleInboundFallback(
	ctx context.Context,
	contact *types.Contact,
	taskID *uuid.UUID,
	task *taskDetails,
	body string,
) {
	normalizedBody := normalizeBody(body)

	isConfirmation := containsAny(normalizedBody, "yes", "confirm", "done", "complete")
	isDelay := containsDelayIndicator(normalizedBody)

	if isConfirmation && a.feedWriter != nil {
		a.writeSubConfirmationCard(ctx, task, contact, body)
	} else if isDelay && a.feedWriter != nil {
		a.writeSubDelayCard(ctx, task, contact, body)
	}
}

// draftFollowUpWithClaude uses Claude to generate a contextual follow-up message
// for a subcontractor who hasn't responded to a confirmation request.
func (a *SubLiaisonAgent) draftFollowUpWithClaude(
	ctx context.Context,
	task *taskDetails,
	contact *types.Contact,
) (string, error) {
	userMessage := fmt.Sprintf(`Draft a follow-up SMS to %s (%s, %s) who hasn't responded to a confirmation request for task "%s" scheduled %s.

Keep it brief (under 160 chars for SMS), professional but friendly, and construction-appropriate.
Use the draft_message tool to create the follow-up.`,
		contact.Name, contact.Company, contact.Role,
		task.Name, formatDate(task.EarlyStart))

	projectCtx := ProjectContext{
		ProjectID: task.ProjectID,
		OrgID:     task.OrgID,
		UserID:    uuid.Nil,
	}

	result, err := a.claudeRunner.Run(ctx, subLiaisonClaudeSystemPrompt, userMessage, projectCtx)
	if err != nil {
		return "", fmt.Errorf("claude follow-up draft failed: %w", err)
	}

	// If Claude produced a text response (the draft), return it
	if result.Text != "" {
		return result.Text, nil
	}

	return "", fmt.Errorf("claude produced no follow-up text")
}

// normalizeBody lowercases and trims a message body for keyword matching.
func normalizeBody(body string) string {
	// Import-free lowercase + trim via existing strings usage in sub_liaison.go
	// This is called from the fallback path only
	b := []byte(body)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}

// containsAny checks if text contains any of the given keywords.
func containsAny(text string, keywords ...string) bool {
	for _, kw := range keywords {
		for i := 0; i <= len(text)-len(kw); i++ {
			if text[i:i+len(kw)] == kw {
				return true
			}
		}
	}
	return false
}

// marshalParsedInbound serializes parsed inbound data for logging.
func marshalParsedInbound(p parsedInbound) string {
	b, err := json.Marshal(p)
	if err != nil {
		return "{}"
	}
	return string(b)
}
