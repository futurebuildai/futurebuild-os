package adapters

import (
	"context"
	"encoding/json"

	"github.com/colton/futurebuild/internal/agents/tools"
	"github.com/colton/futurebuild/internal/service"
)

// ActionToolAdapter bridges service.ActionToolRunner to tools.ActionRunnerAdapter.
// Converts service.ActionScope to tools.Scope for execution.
type ActionToolAdapter struct {
	runner *tools.ActionRunnerAdapter
}

// NewActionToolAdapter creates an adapter satisfying service.ActionToolRunner.
func NewActionToolAdapter(runner *tools.ActionRunnerAdapter) *ActionToolAdapter {
	return &ActionToolAdapter{runner: runner}
}

// ExecuteAction implements service.ActionToolRunner.
func (a *ActionToolAdapter) ExecuteAction(ctx context.Context, actionType string, payload json.RawMessage, scope service.ActionScope) (string, error) {
	return a.runner.ExecuteAction(ctx, actionType, payload, scope.ProjectID, scope.OrgID, scope.UserID)
}

// Compile-time check.
var _ service.ActionToolRunner = (*ActionToolAdapter)(nil)
