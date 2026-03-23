package chat

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/colton/futurebuild/internal/agents"
	"github.com/colton/futurebuild/internal/agents/tools"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
)

// maxHistoryMessages is the max number of prior messages to include for context.
// Keeps Claude within context window while providing conversation continuity.
const maxHistoryMessages = 20

// HistoryLoader loads conversation history for a thread.
// Satisfied by *service.ThreadService.
type HistoryLoader interface {
	GetThreadMessages(ctx context.Context, threadID, projectID, orgID uuid.UUID, limit int) ([]models.ChatMessage, error)
}

// ClaudeOrchestrator replaces the regex-based Orchestrator with Claude-powered
// intent understanding, reasoning, and tool-use execution.
//
// Key differences from Orchestrator:
// - No RegexClassifier — Claude understands natural language directly
// - No createCommand() switch — Claude decides which tools to call
// - Full reasoning partner — helps users think through problems and draft communications
// - Human-in-the-loop — creates approval cards for state-changing actions
type ClaudeOrchestrator struct {
	db             MessagePersister
	runner         *agents.AgentRunner
	streamRunner   *agents.StreamAgentRunner
	history        HistoryLoader

	// System prompt template. ProjectContext is injected at runtime.
	systemPromptBase string
}

// NewClaudeOrchestrator creates a Claude-powered orchestrator.
func NewClaudeOrchestrator(
	persister MessagePersister,
	aiClient ai.Client,
	toolRegistry *tools.Registry,
) *ClaudeOrchestrator {
	return &ClaudeOrchestrator{
		db:               persister,
		runner:           agents.NewAgentRunner(aiClient, toolRegistry),
		systemPromptBase: defaultSystemPrompt,
	}
}

// WithHistory enables conversation history loading for multi-turn chat.
func (o *ClaudeOrchestrator) WithHistory(loader HistoryLoader) *ClaudeOrchestrator {
	o.history = loader
	return o
}

// WithStreamRunner enables streaming chat support.
func (o *ClaudeOrchestrator) WithStreamRunner(sr *agents.StreamAgentRunner) *ClaudeOrchestrator {
	o.streamRunner = sr
	return o
}

// ProcessRequest handles an inbound chat message using Claude.
//
// Flow:
//  1. Persist user message
//  2. Load conversation history for this thread
//  3. Run Claude agent with tools
//  4. Persist assistant response
//  5. Return response with any artifacts
func (o *ClaudeOrchestrator) ProcessRequest(ctx context.Context, userID uuid.UUID, orgID uuid.UUID, req ChatRequest) (*ChatResponse, error) {
	start := time.Now()

	// 1. Persist User Message
	userMsg := models.ChatMessage{
		ID:        uuid.New(),
		ProjectID: req.ProjectID,
		ThreadID:  req.ThreadID,
		UserID:    &userID,
		Role:      types.ChatRoleUser,
		Content:   req.Message,
		CreatedAt: time.Now().UTC(),
	}
	if err := o.db.SaveMessage(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("failed to persist user message: %w", err)
	}

	// 2. Load conversation history and run Claude agent
	projectCtx := agents.ProjectContext{
		ProjectID: req.ProjectID,
		OrgID:     orgID,
		UserID:    userID,
	}

	var result *agents.AgentResult
	var err error

	if o.history != nil {
		// Load prior messages for multi-turn context
		history, histErr := o.history.GetThreadMessages(ctx, req.ThreadID, req.ProjectID, orgID, maxHistoryMessages)
		if histErr != nil {
			slog.Warn("claude orchestrator: failed to load history, proceeding without",
				"thread_id", req.ThreadID, "error", histErr)
		}

		if len(history) > 0 {
			// Convert DB messages to AI messages (exclude the message we just saved)
			msgs := convertToAIMessages(history)
			// Append current user message
			msgs = append(msgs, ai.Message{
				Role:    "user",
				Content: []ai.ContentPart{{Text: req.Message}},
			})
			result, err = o.runner.RunWithHistory(ctx, o.systemPromptBase, msgs, projectCtx)
		} else {
			result, err = o.runner.Run(ctx, o.systemPromptBase, req.Message, projectCtx)
		}
	} else {
		result, err = o.runner.Run(ctx, o.systemPromptBase, req.Message, projectCtx)
	}
	if err != nil {
		slog.Error("claude orchestrator: agent run failed",
			"project_id", req.ProjectID,
			"thread_id", req.ThreadID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)

		// Persist error message so user sees feedback
		errorMsg := models.ChatMessage{
			ID:        uuid.New(),
			ProjectID: req.ProjectID,
			ThreadID:  req.ThreadID,
			UserID:    &userID,
			Role:      types.ChatRoleModel,
			Content:   "I encountered an error processing your request. Please try again.",
			CreatedAt: time.Now().UTC(),
		}
		if persistErr := o.db.SaveMessage(ctx, errorMsg); persistErr != nil {
			slog.Error("claude orchestrator: failed to persist error message", "error", persistErr)
		}
		return nil, fmt.Errorf("agent execution failed: %w", err)
	}

	// 3. Persist assistant response
	modelMsg := models.ChatMessage{
		ID:        uuid.New(),
		ProjectID: req.ProjectID,
		ThreadID:  req.ThreadID,
		UserID:    &userID,
		Role:      types.ChatRoleModel,
		Content:   result.Text,
		CreatedAt: time.Now().UTC(),
	}
	if err := o.db.SaveMessage(ctx, modelMsg); err != nil {
		slog.Error("claude orchestrator: failed to persist model message", "error", err)
		// Non-fatal: return response even if persistence fails
	}

	slog.Info("claude orchestrator: request completed",
		"project_id", req.ProjectID,
		"thread_id", req.ThreadID,
		"turns", result.Turns,
		"tokens", result.TotalTokens,
		"tools_used", result.ToolsUsed,
		"duration_ms", time.Since(start).Milliseconds(),
	)

	// 4. Return response
	return &ChatResponse{
		Reply:  result.Text,
		Intent: types.IntentUnknown, // Claude doesn't classify into discrete intents
	}, nil
}

// ProcessRequestStreaming handles an inbound chat message with streaming response.
// Streams text deltas via the out channel. Persists final response after stream completes.
func (o *ClaudeOrchestrator) ProcessRequestStreaming(ctx context.Context, userID uuid.UUID, orgID uuid.UUID, req ChatRequest, out chan<- ai.StreamChunk) error {
	if o.streamRunner == nil {
		return fmt.Errorf("streaming not configured")
	}

	// Enforce execution timeout to prevent runaway agent loops
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	// 1. Persist User Message
	userMsg := models.ChatMessage{
		ID:        uuid.New(),
		ProjectID: req.ProjectID,
		ThreadID:  req.ThreadID,
		UserID:    &userID,
		Role:      types.ChatRoleUser,
		Content:   req.Message,
		CreatedAt: time.Now().UTC(),
	}
	if err := o.db.SaveMessage(ctx, userMsg); err != nil {
		return fmt.Errorf("failed to persist user message: %w", err)
	}

	// 2. Build message history
	projectCtx := agents.ProjectContext{
		ProjectID: req.ProjectID,
		OrgID:     orgID,
		UserID:    userID,
	}

	var msgs []ai.Message
	if o.history != nil {
		history, histErr := o.history.GetThreadMessages(ctx, req.ThreadID, req.ProjectID, orgID, maxHistoryMessages)
		if histErr != nil {
			slog.Warn("claude stream: failed to load history", "error", histErr)
		}
		if len(history) > 0 {
			msgs = convertToAIMessages(history)
		}
	}
	msgs = append(msgs, ai.Message{
		Role:    "user",
		Content: []ai.ContentPart{{Text: req.Message}},
	})

	// 3. Run streaming agent
	result, err := o.streamRunner.RunStreaming(ctx, o.systemPromptBase, msgs, projectCtx, out)
	if err != nil {
		slog.Error("claude stream: agent run failed",
			"project_id", req.ProjectID,
			"thread_id", req.ThreadID,
			"error", err,
		)
		return fmt.Errorf("streaming agent failed: %w", err)
	}

	// 4. Persist assistant response
	modelMsg := models.ChatMessage{
		ID:        uuid.New(),
		ProjectID: req.ProjectID,
		ThreadID:  req.ThreadID,
		UserID:    &userID,
		Role:      types.ChatRoleModel,
		Content:   result.Text,
		CreatedAt: time.Now().UTC(),
	}
	if err := o.db.SaveMessage(ctx, modelMsg); err != nil {
		slog.Error("claude stream: failed to persist model message", "error", err)
	}

	slog.Info("claude stream: request completed",
		"project_id", req.ProjectID,
		"thread_id", req.ThreadID,
		"turns", result.Turns,
		"tools_used", result.ToolsUsed,
	)

	return nil
}

// OnboardingRequest captures a single turn in the onboarding conversation.
type OnboardingRequest struct {
	SessionID string   `json:"session_id"`
	Message   string   `json:"message"`
	History   []ai.Message `json:"-"` // Prior conversation turns (injected server-side)
}

// OnboardingResponse is the orchestrator's response for onboarding.
type OnboardingResponse struct {
	Reply      string   `json:"reply"`
	ToolsUsed  []string `json:"tools_used,omitempty"`
}

// ProcessOnboardingMessage handles onboarding chat using Claude with onboarding-specific tools.
// Unlike ProcessRequest, this method doesn't require a project scope (pre-creation).
func (o *ClaudeOrchestrator) ProcessOnboardingMessage(
	ctx context.Context,
	systemPrompt string,
	req OnboardingRequest,
) (*OnboardingResponse, error) {
	start := time.Now()

	// Enforce execution timeout to prevent runaway agent loops
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	// Truncate history to prevent context window overflow
	const maxOnboardingHistory = 20
	history := req.History
	if len(history) > maxOnboardingHistory {
		history = history[len(history)-maxOnboardingHistory:]
	}

	// No project context needed — onboarding is pre-project-creation
	projectCtx := agents.ProjectContext{}

	var result *agents.AgentResult
	var err error

	if len(history) > 0 {
		// Multi-turn: include prior conversation
		msgs := make([]ai.Message, len(history))
		copy(msgs, history)
		msgs = append(msgs, ai.Message{
			Role:    "user",
			Content: []ai.ContentPart{{Text: req.Message}},
		})
		result, err = o.runner.RunWithHistory(ctx, systemPrompt, msgs, projectCtx)
	} else {
		result, err = o.runner.Run(ctx, systemPrompt, req.Message, projectCtx)
	}

	if err != nil {
		slog.Error("claude onboarding: agent run failed",
			"session_id", req.SessionID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return nil, fmt.Errorf("onboarding agent failed: %w", err)
	}

	slog.Info("claude onboarding: request completed",
		"session_id", req.SessionID,
		"turns", result.Turns,
		"tools_used", result.ToolsUsed,
		"duration_ms", time.Since(start).Milliseconds(),
	)

	return &OnboardingResponse{
		Reply:     result.Text,
		ToolsUsed: result.ToolsUsed,
	}, nil
}

// convertToAIMessages converts stored ChatMessages to AI message format for Claude.
// Maps ChatRoleUser → "user" and ChatRoleModel → "assistant".
func convertToAIMessages(msgs []models.ChatMessage) []ai.Message {
	result := make([]ai.Message, 0, len(msgs))
	for _, m := range msgs {
		role := "user"
		if m.Role == types.ChatRoleModel {
			role = "assistant"
		}
		result = append(result, ai.Message{
			Role:    role,
			Content: []ai.ContentPart{{Text: m.Content}},
		})
	}
	return result
}

// defaultSystemPrompt is the base system prompt for the chat agent.
// Project-specific context is injected by the AgentRunner via tools.
const defaultSystemPrompt = `You are the AI superintendent for a construction project on FutureBuild.
Your role is to help keep projects on time and on budget.

## Your Capabilities
You have access to tools that let you:
- View project schedules, tasks, and critical path information
- Check procurement status and supply chain risks  
- Look up contacts (subcontractors, project managers)
- Check weather forecasts for the project location
- Draft and send professional communications (email, SMS)
- Create feed cards and approval requests

## How You Work

### As a Thinking Partner
When users bring you problems, reason through them step by step:
- Pull relevant context using your tools before answering
- Analyze situations from multiple angles (schedule impact, budget impact, relationship impact)
- Propose concrete actions, not just advice

### Drafting Communications
When a user needs help composing a message (responding to an upset client, explaining a delay, 
negotiating a change order), use the draft_message tool to compose a professional response.
- Always present the draft for the user to review before sending
- Only use send_email or send_sms AFTER the user explicitly approves the draft
- Match the tone to the situation: empathetic for upset clients, factual for delays, 
  collaborative for negotiations

### Approval Workflow
For ANY action that modifies project state, you MUST use create_approval_card:
- Updating task status or progress
- Sending notifications to contacts  
- Making schedule changes
- Approving change orders or budget adjustments
- Ordering materials

Present your reasoning to the user, then create the approval card. Do NOT execute 
state-changing tools directly without going through the approval flow.

### Non-Standard Situations
For delays, change orders, scope changes, budget overruns, and other non-standard situations:
1. Gather all relevant data (schedule, budget, contacts, history)
2. Analyze the impact on the critical path and budget
3. Present options with pros/cons
4. Create an approval card with your recommendation

## Tone
Be direct, professional, and construction-industry appropriate. You're a superintendent, 
not a chatbot. Be concise but thorough. If you don't have enough information, ask.
`
