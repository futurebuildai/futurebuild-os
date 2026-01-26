package skills

import (
	"context"
	"fmt"

	"github.com/colton/futurebuild/internal/agents"
)

// ProcurementSyncSkill wraps ProcurementAgent.Execute for Tribunal-triggered execution.
// See specs/FUTURESHADE_AGENTS_SPEC.md Section 4.3 (Skill Wrappers)
type ProcurementSyncSkill struct {
	agent *agents.ProcurementAgent
}

// NewProcurementSyncSkill creates a new ProcurementSyncSkill.
func NewProcurementSyncSkill(agent *agents.ProcurementAgent) *ProcurementSyncSkill {
	return &ProcurementSyncSkill{agent: agent}
}

// ID returns the skill identifier.
func (s *ProcurementSyncSkill) ID() string {
	return "procurement_sync"
}

// Execute runs the procurement agent.
func (s *ProcurementSyncSkill) Execute(ctx context.Context, params map[string]any) (Result, error) {
	err := s.agent.Execute(ctx)
	if err != nil {
		return Result{
			Success: false,
			Summary: fmt.Sprintf("Procurement sync failed: %v", err),
		}, err
	}

	return Result{
		Success: true,
		Summary: "Procurement sync completed successfully",
	}, nil
}
