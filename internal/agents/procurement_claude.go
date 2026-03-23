package agents

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

// procurementClaudeSystemPrompt guides Claude for procurement reasoning.
const procurementClaudeSystemPrompt = `You are a construction procurement specialist AI analyzing long-lead items for a project.

Your job: For each procurement item that has reached Warning or Critical status, analyze the situation and create actionable feed cards.

## For each item, you should:
1. Assess the urgency based on lead time, calculated order date, and critical path impact
2. Create an approval card (create_approval_card) recommending the specific action to take:
   - "Order Now" for critical items where the order window is closing
   - "Escalate to PM" for items with vendor issues or significant budget impact
   - "Find Alternative" for items that may be unavailable or overpriced
3. Include the consequence of inaction (e.g., "3-day critical path slip if not ordered by Friday")

## Be specific:
- Include item names, deadlines, and dollar amounts when available
- Recommend specific actions (not just "review this")
- Flag items that affect the critical path as highest priority
`

// writeProcurementCardsWithClaude uses Claude to generate intelligent procurement feed cards.
// Falls back to template-based cards on failure.
func (a *ProcurementAgent) writeProcurementCardsWithClaude(ctx context.Context, results []alertResult) {
	if len(results) == 0 {
		return
	}

	// Build a summary of all items needing attention
	var itemSummary string
	for i, r := range results {
		itemSummary += fmt.Sprintf("%d. %s (Status: %s, Order by: %s, Task: %s)\n",
			i+1, r.ItemName, r.NewStatus, r.CalculatedOrderDate.Format("2006-01-02"), r.ProjectTaskID)
	}

	// Use the first result's project context (all should be same project within a batch)
	projectID := results[0].ProjectID
	orgID := results[0].OrgID

	userMessage := fmt.Sprintf(`Analyze these procurement items that need attention:

%s

Create appropriate approval cards for each item. The most critical items should get priority 0.`, itemSummary)

	projectCtx := ProjectContext{
		ProjectID: projectID,
		OrgID:     orgID,
		UserID:    uuid.Nil, // Agent user
	}

	result, err := a.claudeRunner.Run(ctx, procurementClaudeSystemPrompt, userMessage, projectCtx)
	if err != nil {
		slog.Warn("claude procurement reasoning failed, falling back to template cards",
			"project_id", projectID, "error", err)
		a.writeProcurementCards(ctx, results)
		return
	}

	slog.Info("claude procurement reasoning completed",
		"project_id", projectID,
		"items", len(results),
		"turns", result.Turns,
		"tools_used", result.ToolsUsed,
		"tokens", result.TotalTokens,
	)
}
