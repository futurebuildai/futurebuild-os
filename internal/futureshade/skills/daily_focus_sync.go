package skills

import (
	"context"
	"fmt"

	"github.com/colton/futurebuild/internal/agents"
)

// DailyFocusSyncSkill wraps DailyFocusAgent.Execute for Tribunal-triggered execution.
// See specs/FUTURESHADE_AGENTS_SPEC.md Section 4.3 (Skill Wrappers)
type DailyFocusSyncSkill struct {
	agent *agents.DailyFocusAgent
}

// NewDailyFocusSyncSkill creates a new DailyFocusSyncSkill.
func NewDailyFocusSyncSkill(agent *agents.DailyFocusAgent) *DailyFocusSyncSkill {
	return &DailyFocusSyncSkill{agent: agent}
}

// ID returns the skill identifier.
func (s *DailyFocusSyncSkill) ID() string {
	return "daily_focus_sync"
}

// Execute runs the daily focus agent.
func (s *DailyFocusSyncSkill) Execute(ctx context.Context, params map[string]any) (Result, error) {
	err := s.agent.Execute(ctx)
	if err != nil {
		return Result{
			Success: false,
			Summary: fmt.Sprintf("Daily focus sync failed: %v", err),
		}, err
	}

	return Result{
		Success: true,
		Summary: "Daily focus sync completed successfully",
	}, nil
}
