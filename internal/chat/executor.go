package chat

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
)

// =============================================================================
// COMMAND EXECUTOR (Decorator Pattern)
// =============================================================================
// Wraps command execution with persistence middleware, decoupling:
// - Command logic (what to do)
// - Persistence strategy (how to save)
// - Observability (logging/metrics)
//
// This achieves Single Responsibility Principle compliance in ProcessRequest.
// See PRODUCTION_PLAN.md Orchestrator SRP Refactoring
// =============================================================================

// CommandExecutor wraps command execution with persistence middleware.
// It coordinates command execution with the appropriate persistence strategy
// based on the command's declared ConsistencyLevel.
type CommandExecutor struct {
	strategyRegistry *PersistenceStrategyRegistry
}

// NewCommandExecutor creates a new CommandExecutor with the given strategy registry.
func NewCommandExecutor(registry *PersistenceStrategyRegistry) *CommandExecutor {
	return &CommandExecutor{
		strategyRegistry: registry,
	}
}

// ExecutionContext contains contextual information for command execution.
type ExecutionContext struct {
	UserID    uuid.UUID
	ProjectID uuid.UUID
	Intent    types.Intent
}

// Execute runs the command and persists the result using the appropriate strategy.
// This is the core decorator that:
// 1. Executes the command
// 2. Builds the model message
// 3. Persists via the strategy matching the command's ConsistencyLevel
// 4. Returns the response with intent and artifact
func (e *CommandExecutor) Execute(
	ctx context.Context,
	cmd ChatCommand,
	execCtx ExecutionContext,
) (*ChatResponse, error) {
	// 1. Execute Command with Observability
	cmdStart := time.Now()
	reply, artifact, err := cmd.Execute(ctx)
	cmdDurationMs := time.Since(cmdStart).Milliseconds()

	// Log command execution (Observability: Lane C)
	slog.Info("chat: command executed",
		"intent", execCtx.Intent,
		"project_id", execCtx.ProjectID,
		"duration_ms", cmdDurationMs,
	)

	if err != nil {
		return nil, fmt.Errorf("command execution failed: %w", err)
	}

	// 2. Build Model Message
	modelMsg := models.ChatMessage{
		ID:        uuid.New(),
		ProjectID: execCtx.ProjectID,
		UserID:    execCtx.UserID,
		Role:      types.ChatRoleModel,
		Content:   reply,
		CreatedAt: time.Now().UTC(),
	}

	// 3. Persist via Strategy (Strategy Pattern selection)
	strategy := e.strategyRegistry.Get(cmd.ConsistencyLevel())
	if err := strategy.Persist(ctx, modelMsg); err != nil {
		return nil, err
	}

	// 4. Return Response
	return &ChatResponse{
		Reply:    reply,
		Intent:   execCtx.Intent,
		Artifact: artifact,
	}, nil
}
