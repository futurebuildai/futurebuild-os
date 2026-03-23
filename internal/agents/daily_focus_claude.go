package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/google/uuid"
)

// dailyFocusSystemPrompt is the system prompt for Claude-powered daily focus.
const dailyFocusSystemPrompt = `You are the AI superintendent analyzing a construction project for the daily morning briefing.

Your job: identify the TOP 3-5 things requiring human attention today, ranked by urgency.

## For each item you identify, do one of the following:
1. If it requires a human decision or approval → call create_approval_card with specific recommended action
2. If it's informational but important → call write_feed_card

## Categories to analyze:
- **Weather Impact**: Check the weather forecast. If rain/storms are expected, recommend delays for exterior work or protective measures.
- **Critical Path Tasks**: Tasks on the critical path that start today or this week. Any blockers?
- **Procurement Deadlines**: Items with order deadlines approaching. Recommend ordering if lead time windows are closing.
- **Subcontractor Confirmations**: Subs that haven't confirmed for upcoming tasks. Recommend resending or finding alternatives.
- **Schedule Risks**: Tasks that are falling behind or have dependencies at risk.

## Output Format:
After creating the relevant cards, provide a brief summary paragraph for the daily briefing email.
Keep it direct and construction-industry appropriate. No fluff.

## Important:
- Always use create_approval_card for actions that modify project state
- Be specific in your recommendations (include names, dates, dollar amounts when available)
- Flag the single most critical item as priority 0 (critical)
`

// WithClaudeRunner sets the AgentRunner for Claude-powered reasoning.
// When set, processProject uses Claude to generate actionable approval cards
// in addition to the text briefing.
func (a *DailyFocusAgent) WithClaudeRunner(runner *AgentRunner) *DailyFocusAgent {
	a.claudeRunner = runner
	return a
}

// processProjectWithClaude uses the AgentRunner to generate intelligent
// daily focus cards with actionable recommendations.
func (a *DailyFocusAgent) processProjectWithClaude(ctx context.Context, p models.Project, tasks []models.ProjectTask) (string, error) {
	// Build context message for Claude
	taskSummary := buildTaskSummary(tasks)

	userMessage := fmt.Sprintf(`Analyze today's priorities for project "%s" (ID: %s).
Location: %s
Date: %s

Current tasks:
%s

Please analyze the situation and create the appropriate feed cards for the most important items.
Then provide a brief summary paragraph for the daily briefing email.`,
		p.Name, p.ID, p.Address, a.clock.Now().Format("Monday, Jan 02, 2006"), taskSummary)

	projectCtx := ProjectContext{
		ProjectID: p.ID,
		OrgID:     p.OrgID,
		UserID:    uuid.Nil, // Agent user
	}

	result, err := a.claudeRunner.Run(ctx, dailyFocusSystemPrompt, userMessage, projectCtx)
	if err != nil {
		return "", fmt.Errorf("claude daily focus failed: %w", err)
	}

	slog.Info("claude daily focus completed",
		"project_id", p.ID,
		"turns", result.Turns,
		"tools_used", result.ToolsUsed,
		"tokens", result.TotalTokens,
	)

	return result.Text, nil
}

// buildTaskSummary creates a structured text summary of tasks for Claude.
func buildTaskSummary(tasks []models.ProjectTask) string {
	if len(tasks) == 0 {
		return "No focus tasks identified for today."
	}

	var result string
	for i, t := range tasks {
		critical := ""
		if t.IsOnCriticalPath {
			critical = " [CRITICAL PATH]"
		}
		start := "N/A"
		if t.PlannedStart != nil {
			start = t.PlannedStart.Format("2006-01-02")
		}
		end := "N/A"
		if t.PlannedEnd != nil {
			end = t.PlannedEnd.Format("2006-01-02")
		}

		taskJSON, _ := json.Marshal(map[string]interface{}{
			"id":               t.ID.String(),
			"name":             t.Name,
			"status":           string(t.Status),
			"wbs_code":         t.WBSCode,
			"planned_start":    start,
			"planned_end":      end,
			"is_critical_path": t.IsOnCriticalPath,
			"total_float_days": t.TotalFloatDays,
		})

		result += fmt.Sprintf("%d. %s%s (Status: %s, Start: %s, End: %s)\n   %s\n",
			i+1, t.Name, critical, t.Status, start, end, string(taskJSON))
	}
	return result
}

// writeDailyBriefingCardWithClaude creates a daily briefing card using Claude-generated content.
func (a *DailyFocusAgent) writeDailyBriefingCardWithClaude(ctx context.Context, p models.Project, briefing string, tasks []models.ProjectTask) {
	headline := fmt.Sprintf("Daily Focus: %s — %s", p.Name, a.clock.Now().Format("Jan 02"))

	var critCount int
	for _, t := range tasks {
		if t.IsOnCriticalPath {
			critCount++
		}
	}
	var consequence *string
	if critCount > 0 {
		c := fmt.Sprintf("%d critical-path task(s) active today", critCount)
		consequence = &c
	}

	agentSource := "DailyFocusAgent"
	endOfDay := time.Date(
		a.clock.Now().Year(), a.clock.Now().Month(), a.clock.Now().Day(),
		23, 59, 59, 0, a.clock.Now().Location(),
	)

	card := &models.FeedCard{
		OrgID:       p.OrgID,
		ProjectID:   p.ID,
		CardType:    models.FeedCardDailyBriefing,
		Priority:    models.FeedCardPriorityNormal,
		Headline:    headline,
		Body:        briefing,
		Consequence: consequence,
		Horizon:     models.FeedCardHorizonToday,
		AgentSource: &agentSource,
		ExpiresAt:   &endOfDay,
		Actions: []models.FeedCardAction{
			{ID: "view_briefing", Label: "View Project", Style: "primary"},
			{ID: "dismiss", Label: "Dismiss", Style: "secondary"},
		},
	}

	if err := a.feedWriter.WriteCard(ctx, card); err != nil {
		slog.Error("failed to write claude daily briefing feed card", "project_id", p.ID, "error", err)
	}
}
