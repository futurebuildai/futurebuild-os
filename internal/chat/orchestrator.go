package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/platform/db"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// --- Tool Interfaces (The "Limbs") ---
// See PRODUCTION_PLAN.md Step 43.3

// TaskService defines operations for task status updates.
type TaskService interface {
	UpdateTaskStatus(ctx context.Context, taskID, projectID, orgID uuid.UUID, status types.TaskStatus) error
}

// ScheduleService defines operations for schedule retrieval and recalculation.
type ScheduleService interface {
	GetTask(ctx context.Context, taskID, projectID, orgID uuid.UUID) (*models.ProjectTask, error)
	GetProjectSchedule(ctx context.Context, projectID, orgID uuid.UUID) (*service.ProjectScheduleSummary, error)
}

// InvoiceService defines operations for invoice processing.
type InvoiceService interface {
	AnalyzeInvoice(ctx context.Context, orgID uuid.UUID, docID uuid.UUID) (uuid.UUID, *types.InvoiceExtraction, error)
	SaveExtraction(ctx context.Context, projectID uuid.UUID, extraction *types.InvoiceExtraction, sourceDocID *uuid.UUID) (uuid.UUID, error)
}

// MessagePersister defines operations for saving chat messages.
type MessagePersister interface {
	SaveMessage(ctx context.Context, msg models.ChatMessage) error
	// Pool returns the underlying Transactor for callers that need to start transactions.
	// See PRODUCTION_PLAN.md Step 45 (Zombie Write Fix)
	Pool() Transactor
}

// --- Orchestrator Struct ---

// Orchestrator is the central traffic controller for the chat system.
// See PRODUCTION_PLAN.md Step 43.3
// Critical Blocker C Remediation: Added dlq for async retry of failed messages
// P0 FIX (Blocker B): DLQ is now MANDATORY for compliance audit trails.
// Audit Trail Durability: Added WAL and circuit breaker for ACID compliance.
type Orchestrator struct {
	db              MessagePersister
	classifier      IntentClassifier
	TaskService     TaskService
	ScheduleService ScheduleService
	InvoiceService  InvoiceService
	dlq             DLQPersister        // REQUIRED - must never be nil
	wal             AuditWAL            // Local fallback for DLQ failures (optional)
	circuitBreaker  AuditCircuitBreaker // Controls read-only mode (optional)
}

// --- Command Pattern (The "Actions") ---
// See PRODUCTION_PLAN.md Step 43 (Command Pattern Refactor)
// See PRODUCTION_PLAN.md Step 44 (Artifact return for Rich UI)

// ChatCommand defines the interface for intent-specific execution.
// Returns text reply, optional artifact for Rich UI, and error.
type ChatCommand interface {
	Execute(ctx context.Context) (string, *Artifact, error)
}

// GetScheduleCommand retrieves project schedule summary from ScheduleService.
type GetScheduleCommand struct {
	scheduleService ScheduleService
	projectID       uuid.UUID
	orgID           uuid.UUID
}

// Execute calls the ScheduleService and formats the response.
// Returns text summary, schedule_view artifact with structured data, and error.
// See PRODUCTION_PLAN.md Step 44 (Artifact Mapping)
func (c *GetScheduleCommand) Execute(ctx context.Context) (string, *Artifact, error) {
	summary, err := c.scheduleService.GetProjectSchedule(ctx, c.projectID, c.orgID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get schedule: %w", err)
	}

	// ENGINEERING STANDARD: Defensive nil check per Operation Ironclad Task 6
	// Prevents nil pointer dereference even if service returns (nil, nil)
	if summary == nil {
		return "", nil, fmt.Errorf("internal error: schedule summary is nil")
	}

	// Build text response for chat display
	reply := fmt.Sprintf(
		"Project End Date: %s\nCritical Path Tasks: %d\nTotal Tasks: %d\nCompleted: %d",
		summary.ProjectEnd.Format("Jan 02, 2006"),
		summary.CriticalPathCount,
		summary.TotalTasks,
		summary.CompletedTasks,
	)

	// Build Rich UI artifact with structured data
	// See PRODUCTION_PLAN.md Step 44 (schedule_view artifact)
	artifactData, err := json.Marshal(summary)
	if err != nil {
		// Non-fatal: return text without artifact on serialization failure
		return reply, nil, nil
	}

	artifact := &Artifact{
		Type:  ArtifactTypeScheduleView,
		Title: "Project Schedule Summary",
		Data:  artifactData,
	}

	return reply, artifact, nil
}

// PlaceholderCommand returns a static message for unimplemented intents.
type PlaceholderCommand struct {
	message string
}

// Execute returns the placeholder message with no artifact.
// Placeholder commands do not produce Rich UI components.
func (c *PlaceholderCommand) Execute(_ context.Context) (string, *Artifact, error) {
	return c.message, nil, nil
}

// NewOrchestrator creates a new Orchestrator with injected dependencies.
// Uses the default RegexClassifier for intent classification.
// See PRODUCTION_PLAN.md Step 43.3
// ENGINEERING STANDARD: Accepts MessagePersister interface, not *pgxpool.Pool,
// to enable strict dependency injection and testability.
// P0 FIX (Blocker B): DLQ is now REQUIRED. Panics on nil to fail fast at startup.
// Audit Trail Durability: WAL and circuit breaker are optional but recommended.
func NewOrchestrator(
	persister MessagePersister,
	taskService TaskService,
	scheduleService ScheduleService,
	invoiceService InvoiceService,
	dlq DLQPersister,
	wal AuditWAL,
	circuitBreaker AuditCircuitBreaker,
) *Orchestrator {
	if dlq == nil {
		panic("chat: DLQPersister is required for compliance audit trails")
	}
	return &Orchestrator{
		db:              persister,
		classifier:      NewDefaultRegexClassifier(),
		TaskService:     taskService,
		ScheduleService: scheduleService,
		InvoiceService:  invoiceService,
		dlq:             dlq,
		wal:             wal,
		circuitBreaker:  circuitBreaker,
	}
}

// NewOrchestratorWithPersister creates a new Orchestrator with a custom MessagePersister and IntentClassifier.
// This is primarily used for testing to inject mock dependencies.
// See PRODUCTION_PLAN.md Step 43.4
// P0 FIX (Blocker B): DLQ is now REQUIRED. Panics on nil to fail fast at startup.
// Audit Trail Durability: WAL and circuit breaker are optional but recommended.
func NewOrchestratorWithPersister(
	persister MessagePersister,
	classifier IntentClassifier,
	taskService TaskService,
	scheduleService ScheduleService,
	invoiceService InvoiceService,
	dlq DLQPersister,
	wal AuditWAL,
	circuitBreaker AuditCircuitBreaker,
) *Orchestrator {
	if dlq == nil {
		panic("chat: DLQPersister is required for compliance audit trails")
	}
	return &Orchestrator{
		db:              persister,
		classifier:      classifier,
		TaskService:     taskService,
		ScheduleService: scheduleService,
		InvoiceService:  invoiceService,
		dlq:             dlq,
		wal:             wal,
		circuitBreaker:  circuitBreaker,
	}
}

// ProcessRequest is the main entry point for handling a user's chat message.
// Implements a Two-Lane Consistency Strategy:
//   - Lane A (Slow/External): AI operations with best-effort persistence
//   - Lane B (Fast/Internal): DB operations with strict consistency
//
// See PRODUCTION_PLAN.md Step 43.3, 43.4, Step 2 (Hybrid Consistency)
func (o *Orchestrator) ProcessRequest(ctx context.Context, userID uuid.UUID, orgID uuid.UUID, req ChatRequest) (*ChatResponse, error) {
	// 1. Inbound Persistence: Save User Message (ALWAYS required)
	// See DATA_SPINE_SPEC.md Section 5.3
	userMsg := models.ChatMessage{
		ID:        uuid.New(),
		ProjectID: req.ProjectID,
		UserID:    userID,
		Role:      types.ChatRoleUser,
		Content:   req.Message,
		CreatedAt: time.Now().UTC(),
	}
	if err := o.db.SaveMessage(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("failed to persist user message: %w", err)
	}

	// 2. Classify Intent (using injected classifier)
	// See PRODUCTION_PLAN.md Step 43.2 (Weighted Regex Router)
	intent := o.classifier.Classify(req.Message)

	// 3. Route & Execute Command with Observability
	// See PRODUCTION_PLAN.md Step 43 (Command Pattern)
	// See PRODUCTION_PLAN.md Step 44 (Artifact return for Rich UI)
	cmdStart := time.Now()
	reply, artifact, err := o.routeIntent(ctx, intent, req.ProjectID, orgID)
	cmdDurationMs := time.Since(cmdStart).Milliseconds()

	// Log command execution (Lane C: Observability)
	slog.Info("chat: command executed",
		"intent", intent,
		"project_id", req.ProjectID,
		"duration_ms", cmdDurationMs,
	)

	if err != nil {
		return nil, fmt.Errorf("command execution failed: %w", err)
	}

	// 4. Outbound Persistence: Save Model Response
	// Two-Lane Strategy: Differentiate error handling by intent type
	// See DATA_SPINE_SPEC.md Section 5.3, Step 2 (Hybrid Consistency)
	modelMsg := models.ChatMessage{
		ID:        uuid.New(),
		ProjectID: req.ProjectID,
		UserID:    userID,
		Role:      types.ChatRoleModel,
		Content:   reply,
		CreatedAt: time.Now().UTC(),
	}

	if err := o.db.SaveMessage(ctx, modelMsg); err != nil {
		// Two-Lane Decision Point:
		// Lane A (Slow/External): AI operations succeed, persistence failure is non-fatal
		// Lane B (Fast/Internal): DB operations require strict consistency
		if isSlowExternalIntent(intent) {
			// Audit Trail Durability: Check circuit breaker first
			// If open, reject new Lane A writes (read-only mode)
			if o.circuitBreaker != nil && o.circuitBreaker.IsOpen() {
				slog.Error("AUDIT SYSTEM UNAVAILABLE: Degrading to read-only mode",
					"intent", intent,
					"project_id", req.ProjectID,
					"user_id", userID,
				)
				return nil, fmt.Errorf("system temporarily unavailable: audit system degraded")
			}

			// Lane A: Action succeeded but chat history save failed
			// CRITICAL: Log ERROR but attempt fallback persistence
			// The expensive AI operation completed - we must persist the audit trail
			slog.Error("CRITICAL: Action succeeded but chat history save failed",
				"intent", intent,
				"project_id", req.ProjectID,
				"user_id", userID,
				"error", err,
			)

			// Fallback 1: Try DLQ (Redis-backed async retry)
			if dlqErr := o.dlq.EnqueueRetry(ctx, modelMsg); dlqErr != nil {
				slog.Error("CRITICAL COMPLIANCE FAILURE: DLQ enqueue failed - trying WAL",
					"message_id", modelMsg.ID,
					"error", dlqErr,
				)

				// Fallback 2: Try local WAL (disk-based last resort)
				if o.wal != nil {
					walErr := o.wal.AppendRecord(ctx, modelMsg)
					if walErr == nil {
						slog.Warn("Audit record written to local WAL for recovery",
							"message_id", modelMsg.ID,
						)
						if o.circuitBreaker != nil {
							o.circuitBreaker.RecordSuccess()
						}
						// WAL succeeded - return success to user
						return &ChatResponse{
							Reply:    reply,
							Intent:   intent,
							Artifact: artifact,
						}, nil
					}
					slog.Error("CATASTROPHIC: WAL write also failed",
						"message_id", modelMsg.ID,
						"error", walErr,
					)
				}

				// All audit systems failed - trip circuit breaker and return error
				if o.circuitBreaker != nil {
					o.circuitBreaker.RecordFailure()
				}
				// Return error to user - unaudited action cannot proceed
				// See Audit Trail Durability remediation
				return nil, fmt.Errorf("audit system unavailable: please try again later")
			}

			// DLQ succeeded
			slog.Info("Message enqueued to DLQ for retry",
				"message_id", modelMsg.ID,
			)
			if o.circuitBreaker != nil {
				o.circuitBreaker.RecordSuccess()
			}
			// Return success - user sees the result, audit will be retried async
			return &ChatResponse{
				Reply:    reply,
				Intent:   intent,
				Artifact: artifact,
			}, nil
		}

		// Lane B: Strict consistency - propagate error for fast/internal ops
		slog.Error("chat: model persistence failed (strict mode)",
			"intent", intent,
			"project_id", req.ProjectID,
			"user_id", userID,
			"error", err,
		)
		return nil, fmt.Errorf("failed to persist model message: %w", err)
	}

	// 5. Return Response (with optional artifact for Rich UI)
	// See PRODUCTION_PLAN.md Step 44
	return &ChatResponse{
		Reply:    reply,
		Intent:   intent,
		Artifact: artifact,
	}, nil
}

// isSlowExternalIntent returns true for intents that involve slow/external operations
// (AI, Vision, LLM calls) where at-least-once execution with best-effort persistence is acceptable.
// Returns false for fast/internal intents (DB operations) that require strict consistency.
// See Step 2: Two-Lane Consistency Strategy
func isSlowExternalIntent(intent types.Intent) bool {
	switch intent {
	case types.IntentProcessInvoice,
		types.IntentExplainDelay,
		types.IntentUnknown:
		// Lane A: Slow/External - AI/Vision operations
		// These are expensive; if they succeed, we must return success to user
		return true
	case types.IntentGetSchedule,
		types.IntentUpdateTaskStatus:
		// Lane B: Fast/Internal - DB operations
		// These are idempotent or read-only; retry is safe, strict consistency required
		return false
	default:
		// Unknown intents default to Lane A (graceful degradation)
		return true
	}
}

// routeIntent creates and executes the appropriate command for the given intent.
// See PRODUCTION_PLAN.md Step 43 (Command Pattern)
// See PRODUCTION_PLAN.md Step 44 (Artifact return for Rich UI)
func (o *Orchestrator) routeIntent(ctx context.Context, intent types.Intent, projectID, orgID uuid.UUID) (string, *Artifact, error) {
	cmd := o.createCommand(intent, projectID, orgID)
	return cmd.Execute(ctx)
}

// createCommand is the Command Factory that instantiates the correct command.
func (o *Orchestrator) createCommand(intent types.Intent, projectID, orgID uuid.UUID) ChatCommand {
	switch intent {
	case types.IntentGetSchedule:
		return &GetScheduleCommand{
			scheduleService: o.ScheduleService,
			projectID:       projectID,
			orgID:           orgID,
		}
	case types.IntentProcessInvoice:
		return &PlaceholderCommand{message: "I can help you process that invoice. Please upload the document."}
	case types.IntentExplainDelay:
		return &PlaceholderCommand{message: "I'm analyzing the current schedule to explain potential delays."}
	case types.IntentUpdateTaskStatus:
		return &PlaceholderCommand{message: "Ready to update the task status. Please confirm the task and new status."}
	case types.IntentContactSubcontractor:
		// See PRODUCTION_PLAN.md Step 47 (Sub Liaison Agent)
		// Full SubLiaisonCommand implementation requires task parsing and agent injection.
		// Placeholder returns guidance for now; full wiring is in WebhookHandler.
		return &PlaceholderCommand{message: "I'll reach out to the subcontractor for you. Which task would you like me to confirm arrival for?"}
	default:
		return &PlaceholderCommand{message: "I'm not sure how to help with that. Could you please rephrase your request?"}
	}
}

// --- Default MessagePersister Implementation ---

// TxExecutor defines the interface for transaction-capable execution.
// This is used by the transactional SaveMessage to insert into both
// chat_messages and chat_tool_usage atomically.
type TxExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// Transactor defines the interface for beginning transactions.
// *pgxpool.Pool satisfies this interface.
type Transactor interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// DBExecutor defines a subset of pgxpool.Pool methods needed for execution.
// This allows us to mock the database connection for 100% coverage.
// DEPRECATED: Use Transactor for new code requiring transactions.
type DBExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
}

// PgxMessageStore implements MessagePersister using a Transactor (pgxpool or mock).
// Phase 2 Remediation: Now supports transactional writes for ACID compliance.
type PgxMessageStore struct {
	pool Transactor
}

// NewPgxMessageStore creates a new PgxMessageStore.
// The pool must satisfy the Transactor interface (*pgxpool.Pool does).
func NewPgxMessageStore(pool Transactor) *PgxMessageStore {
	return &PgxMessageStore{pool: pool}
}

// Pool returns the underlying Transactor for transaction management.
// This allows callers to start transactions and inject them into context.
// See PRODUCTION_PLAN.md Step 45 (Zombie Write Fix)
func (s *PgxMessageStore) Pool() Transactor {
	return s.pool
}

// SaveMessage persists a ChatMessage and its ToolCalls to the database atomically.
// CRITICAL FAANG STANDARD: Both chat_messages and chat_tool_usage inserts
// happen in a single transaction. If tool usage save fails, message save rolls back.
// See DATA_SPINE_SPEC.md Section 5.3, Phase 2 Remediation Task 1.
//
// Distributed Transaction Support (Step 45 Zombie Write Fix):
//   - If a transaction is injected via context (db.InjectTx), uses that transaction.
//     Caller owns Commit/Rollback lifecycle.
//   - Otherwise, starts its own transaction (legacy behavior).
func (s *PgxMessageStore) SaveMessage(ctx context.Context, msg models.ChatMessage) error {
	// Check for context-propagated transaction (caller owns Tx lifecycle)
	if tx, ok := db.ExtractTx(ctx); ok {
		return s.saveMessageWithTx(ctx, tx, msg)
	}

	// Legacy behavior: start own transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Defer rollback - no-op if already committed
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := s.saveMessageWithTx(ctx, tx, msg); err != nil {
		return err
	}

	// Commit transaction (we own it)
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// saveMessageWithTx performs the actual message save using the provided transaction.
// This is the internal implementation shared by both injected and local transactions.
func (s *PgxMessageStore) saveMessageWithTx(ctx context.Context, tx pgx.Tx, msg models.ChatMessage) error {
	// 1. Insert into chat_messages
	messageQuery := `
		INSERT INTO chat_messages (id, project_id, user_id, role, content, tool_calls, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := tx.Exec(ctx, messageQuery,
		msg.ID, msg.ProjectID, msg.UserID, msg.Role, msg.Content, msg.ToolCalls, msg.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("db insert chat_messages failed: %w", err)
	}

	// 2. Insert each ToolCall into chat_tool_usage (normalized table)
	if len(msg.ToolCalls) > 0 {
		toolQuery := `
			INSERT INTO chat_tool_usage (message_id, tool_name, input_payload, output_payload, created_at)
			VALUES ($1, $2, $3, $4, $5)
		`
		for _, tc := range msg.ToolCalls {
			_, err = tx.Exec(ctx, toolQuery,
				msg.ID,        // message_id FK
				tc.Name,       // tool_name
				tc.Args,       // input_payload (JSONB)
				tc.Response,   // output_payload (JSONB stored as string, will be cast)
				msg.CreatedAt, // created_at
			)
			if err != nil {
				return fmt.Errorf("db insert chat_tool_usage failed for tool %s: %w", tc.Name, err)
			}
		}
	}

	return nil
}
