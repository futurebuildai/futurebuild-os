package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// ActionRunnerAdapter bridges the service.ActionToolRunner interface with the tools.Registry.
// When an agent action is approved, this adapter injects the correct scope and executes
// the stored tool call through the registry.
type ActionRunnerAdapter struct {
	registry *Registry
}

// NewActionRunnerAdapter creates an adapter that wraps a tool registry.
func NewActionRunnerAdapter(registry *Registry) *ActionRunnerAdapter {
	return &ActionRunnerAdapter{registry: registry}
}

// ActionScope mirrors service.ActionScope to avoid import cycles.
// The server wiring layer handles the conversion.
type ActionScope struct {
	ProjectID uuid.UUID
	OrgID     uuid.UUID
	UserID    uuid.UUID
}

// ExecuteAction runs a tool by name with the given payload and scope.
// This satisfies the service.ActionToolRunner interface (via adapter in server.go).
func (a *ActionRunnerAdapter) ExecuteAction(ctx context.Context, actionType string, payload json.RawMessage, projectID, orgID, userID uuid.UUID) (string, error) {
	// Inject scope into context
	scopedCtx := WithScope(ctx, Scope{
		ProjectID: projectID,
		OrgID:     orgID,
		UserID:    userID,
	})

	// Execute the tool
	output, err := a.registry.Execute(scopedCtx, actionType, payload)
	if err != nil {
		return "", fmt.Errorf("execute approved action %q: %w", actionType, err)
	}
	return output, nil
}
