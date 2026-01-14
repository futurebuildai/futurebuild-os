package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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
}

// InvoiceService defines operations for invoice processing.
type InvoiceService interface {
	AnalyzeInvoice(ctx context.Context, orgID uuid.UUID, docID uuid.UUID) (uuid.UUID, *types.InvoiceExtraction, error)
	SaveExtraction(ctx context.Context, projectID uuid.UUID, extraction *types.InvoiceExtraction, sourceDocID *uuid.UUID) (uuid.UUID, error)
}

// MessagePersister defines operations for saving chat messages.
type MessagePersister interface {
	SaveMessage(ctx context.Context, msg models.ChatMessage) error
}

// --- Orchestrator Struct ---

// Orchestrator is the central traffic controller for the chat system.
// See PRODUCTION_PLAN.md Step 43.3
type Orchestrator struct {
	db              MessagePersister
	TaskService     TaskService
	ScheduleService ScheduleService
	InvoiceService  InvoiceService
}

// NewOrchestrator creates a new Orchestrator with injected dependencies.
// See PRODUCTION_PLAN.md Step 43.3
func NewOrchestrator(
	db *pgxpool.Pool,
	taskService TaskService,
	scheduleService ScheduleService,
	invoiceService InvoiceService,
) *Orchestrator {
	return &Orchestrator{
		db:              &PgxMessageStore{pool: db},
		TaskService:     taskService,
		ScheduleService: scheduleService,
		InvoiceService:  invoiceService,
	}
}

// ProcessRequest is the main entry point for handling a user's chat message.
// Flow: Persist(User) -> Classify -> Route -> Persist(Model) -> Return
// See PRODUCTION_PLAN.md Step 43.3
func (o *Orchestrator) ProcessRequest(ctx context.Context, userID uuid.UUID, req ChatRequest) (*ChatResponse, error) {
	// 1. Inbound Persistence: Save User Message
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

	// 2. Classify Intent
	// See PRODUCTION_PLAN.md Step 43.2
	intent := ClassifyIntent(req.Message)

	// 3. Route & Generate Reply (V1 Placeholder Logic)
	reply := o.routeIntent(ctx, intent)

	// 4. Outbound Persistence: Save Model Response
	// See DATA_SPINE_SPEC.md Section 5.3
	modelMsg := models.ChatMessage{
		ID:        uuid.New(),
		ProjectID: req.ProjectID,
		UserID:    userID,
		Role:      types.ChatRoleModel,
		Content:   reply,
		CreatedAt: time.Now().UTC(),
	}
	if err := o.db.SaveMessage(ctx, modelMsg); err != nil {
		return nil, fmt.Errorf("failed to persist model message: %w", err)
	}

	// 5. Return Response
	return &ChatResponse{
		Reply:  reply,
		Intent: intent,
	}, nil
}

// routeIntent maps an Intent to a placeholder response string.
// In later steps, this will dispatch to actual service logic.
func (o *Orchestrator) routeIntent(_ context.Context, intent types.Intent) string {
	switch intent {
	case types.IntentProcessInvoice:
		return "I can help you process that invoice. Please upload the document."
	case types.IntentExplainDelay:
		return "I'm analyzing the current schedule to explain potential delays."
	case types.IntentGetSchedule:
		return "Fetching the latest schedule information for your project."
	case types.IntentUpdateTaskStatus:
		return "Ready to update the task status. Please confirm the task and new status."
	default:
		return "I'm not sure how to help with that. Could you please rephrase your request?"
	}
}

// --- Default MessagePersister Implementation ---

// PgxMessageStore implements MessagePersister using pgxpool.
type PgxMessageStore struct {
	pool *pgxpool.Pool
}

// SaveMessage persists a ChatMessage to the database.
// See DATA_SPINE_SPEC.md Section 5.3
func (s *PgxMessageStore) SaveMessage(ctx context.Context, msg models.ChatMessage) error {
	query := `
		INSERT INTO chat_messages (id, project_id, user_id, role, content, tool_calls, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := s.pool.Exec(ctx, query,
		msg.ID, msg.ProjectID, msg.UserID, msg.Role, msg.Content, msg.ToolCalls, msg.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("db insert failed: %w", err)
	}
	return nil
}
