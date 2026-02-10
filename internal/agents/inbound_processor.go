// Package agents provides AI-powered business logic components for FutureBuild.
package agents

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/clock"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Package-Level Compiled Regexes ---
// See PRODUCTION_PLAN.md Step 48 (Performance Optimization)
var (
	inboundPercentageRe = regexp.MustCompile(`^(\d{1,3})\s*%$`)
	inboundImageURLRe   = regexp.MustCompile(`https?://[^\s]+\.(jpg|jpeg|png|gif|webp)`)
)

// InboundMessage represents a normalized webhook payload from Twilio/SendGrid.
// See PRODUCTION_PLAN.md Step 48 (Payload Normalization)
type InboundMessage struct {
	ExternalID string   // Provider MessageSid (Twilio) or Message-ID (SendGrid) for idempotency
	Sender     string   // Phone number or email address
	Body       string   // Message content
	ImageURLs  []string // Extracted image URLs from body
	Channel    string   // "SMS" or "Email"
}

// InboundContactLookup resolves a sender identifier to a Contact.
// See PRODUCTION_PLAN.md Step 48 (Identity Resolution)
type InboundContactLookup interface {
	FindContactBySender(ctx context.Context, sender string) (*types.Contact, error)
}

// InboundContextResolver finds the task context from recent communication.
// See PRODUCTION_PLAN.md Step 48 (Context Binding)
type InboundContextResolver interface {
	FindRecentOutboundTask(ctx context.Context, contactID uuid.UUID, window time.Duration) (*uuid.UUID, *uuid.UUID, error)
}

// InboundProgressUpdater handles task progress updates and schedule recalculation.
// See PRODUCTION_PLAN.md Step 48 (State Machine Integration)
type InboundProgressUpdater interface {
	UpdateTaskProgress(ctx context.Context, taskID uuid.UUID, percent int) error
	RecalculateSchedule(ctx context.Context, projectID, orgID uuid.UUID) error
}

// InboundVisionVerifier triggers AI vision verification for site photos.
// See PRODUCTION_PLAN.md Step 48 (Vision Trigger)
type InboundVisionVerifier interface {
	VerifyTask(ctx context.Context, imageURL, taskDescription string) (bool, float64, error)
}

// InboundProcessor handles incoming webhook messages from subcontractors.
// This is a Reactor (inbound/passive), separate from SubLiaisonAgent (outbound/active).
// See PRODUCTION_PLAN.md Step 48 (Separation of Concerns Amendment)
type InboundProcessor struct {
	db         *pgxpool.Pool
	directory  InboundContactLookup
	schedule   InboundProgressUpdater
	vision     InboundVisionVerifier
	log        *slog.Logger
	clock      clock.Clock
	feedWriter FeedWriter // V2: writes sub confirmation/delay cards to portfolio feed
}

// NewInboundProcessor creates a new InboundProcessor with injected dependencies.
// See PRODUCTION_PLAN.md Step 48
// Refactored for deterministic simulation: PRODUCTION_PLAN.md Step 49
func NewInboundProcessor(
	db *pgxpool.Pool,
	directory InboundContactLookup,
	schedule InboundProgressUpdater,
	vision InboundVisionVerifier,
	clk clock.Clock,
) *InboundProcessor {
	return &InboundProcessor{
		db:        db,
		directory: directory,
		schedule:  schedule,
		vision:    vision,
		log:       slog.With("component", "inbound_processor"),
		clock:     clk,
	}
}

// WithFeedWriter sets the feed writer for V2 portfolio feed card generation.
// If not set, no feed cards are written (backward compatible).
func (p *InboundProcessor) WithFeedWriter(fw FeedWriter) *InboundProcessor {
	p.feedWriter = fw
	return p
}

// ProcessIncoming handles a normalized inbound message.
// See PRODUCTION_PLAN.md Step 48
//
// Logic Flow:
// 1. Idempotency Check (external_id unique constraint)
// 2. Identity Resolution (sender → contact)
// 3. Context Binding (find task from recent outbound)
// 4. Intent Parsing (percentage → progress, confirmation → ACK, image → vision)
// 5. State Machine (100% → trigger CPM recalculation)
func (p *InboundProcessor) ProcessIncoming(ctx context.Context, msg InboundMessage) error {
	// Step 1: Idempotency Check
	// See L7 Amendment: Database-level idempotency via external_id unique index
	if msg.ExternalID != "" {
		exists, err := p.checkExternalIDExists(ctx, msg.ExternalID)
		if err != nil {
			p.log.Error("idempotency check failed", "error", err, "external_id", msg.ExternalID)
			// Continue processing - fail open for availability
		} else if exists {
			p.log.Info("duplicate message detected, skipping",
				"external_id", msg.ExternalID,
				"sender", msg.Sender,
			)
			return nil
		}
	}

	// Step 2: Identity Resolution
	contact, err := p.directory.FindContactBySender(ctx, msg.Sender)
	if err != nil {
		p.log.Warn("unknown sender",
			"sender", msg.Sender,
			"channel", msg.Channel,
			"error", err,
		)
		// Log the message for audit but don't process
		_ = p.logInboundMessage(ctx, nil, nil, msg, "UNKNOWN_SENDER")
		return nil // Not an error - expected for unknown callers
	}

	// Step 3: Context Binding
	// Find the most recent outbound message to this contact to infer task context
	taskID, projectID, err := p.findRecentOutboundContext(ctx, contact.ID, 48*time.Hour)
	if err != nil {
		p.log.Warn("no recent outbound context for sender",
			"contact_id", contact.ID,
			"sender", msg.Sender,
		)
		_ = p.logInboundMessage(ctx, &contact.ID, nil, msg, "NO_CONTEXT")
		return nil
	}

	// Step 4: Intent Parsing
	normalizedBody := strings.ToLower(strings.TrimSpace(msg.Body))

	// Priority 1: Check for percentage update
	// P0 Fix: Transactional idempotency - lock via log insert BEFORE side-effects
	// See PRODUCTION_PLAN.md Phase 49 Retrofit (Operation Ironclad Task 1)
	if percent, ok := p.parsePercentage(normalizedBody); ok {
		if err := p.handleProgressUpdateTransactional(ctx, *taskID, *projectID, contact, percent, msg); err != nil {
			p.log.Error("failed to process progress update",
				"task_id", taskID,
				"percent", percent,
				"error", err,
			)
		}
		return nil
	}

	// Priority 2: Check for image URL (Vision trigger)
	if len(msg.ImageURLs) > 0 {
		if err := p.handleVisionVerification(ctx, *taskID, contact, msg); err != nil {
			p.log.Error("vision verification failed",
				"task_id", taskID,
				"image_count", len(msg.ImageURLs),
				"error", err,
			)
		}
		return nil
	}

	// Priority 3: Check for confirmation keywords
	if p.isConfirmation(normalizedBody) {
		_ = p.logInboundMessage(ctx, &contact.ID, taskID, msg, "ACK")
		p.log.Info("confirmation received",
			"contact_id", contact.ID,
			"task_id", taskID,
		)
		// V2 Feed: Write sub_confirmation card
		if p.feedWriter != nil && taskID != nil {
			p.writeInboundConfirmationCard(ctx, *taskID, *projectID, contact, msg.Body)
		}
		return nil
	}

	// Priority 4: Check for delay/issue indicators
	if p.isDelayIndicator(normalizedBody) {
		if err := p.createRiskFlag(ctx, *taskID, msg.Body); err != nil {
			p.log.Error("failed to create risk flag", "error", err)
		}
		_ = p.logInboundMessage(ctx, &contact.ID, taskID, msg, "DELAY_FLAGGED")
		// V2 Feed: Write sub delay card (elevated priority)
		if p.feedWriter != nil && taskID != nil {
			p.writeInboundDelayCard(ctx, *taskID, *projectID, contact, msg.Body)
		}
		return nil
	}

	// Default: Log as general message
	_ = p.logInboundMessage(ctx, &contact.ID, taskID, msg, "GENERAL")
	return nil
}

// handleProgressUpdate processes a percentage update and triggers CPM recalc if needed.
// See PRODUCTION_PLAN.md Step 48 (State Machine Integration)
// DEPRECATED: Use handleProgressUpdateTransactional for idempotent processing.
func (p *InboundProcessor) handleProgressUpdate(
	ctx context.Context,
	taskID, projectID uuid.UUID,
	contact *types.Contact,
	percent int,
	msg InboundMessage,
) error {
	// Update task progress
	if err := p.schedule.UpdateTaskProgress(ctx, taskID, percent); err != nil {
		return fmt.Errorf("update task progress: %w", err)
	}

	p.log.Info("task progress updated via inbound message",
		"task_id", taskID,
		"percent", percent,
		"sender", msg.Sender,
		"channel", msg.Channel,
	)

	// Step 5: State Machine - 100% triggers CPM recalculation
	if percent == 100 {
		p.log.Info("task marked complete, triggering schedule recalculation",
			"task_id", taskID,
			"project_id", projectID,
		)

		// Fetch org_id for the project
		orgID, err := p.getProjectOrgID(ctx, projectID)
		if err != nil {
			return fmt.Errorf("get project org_id: %w", err)
		}

		if err := p.schedule.RecalculateSchedule(ctx, projectID, orgID); err != nil {
			return fmt.Errorf("recalculate schedule: %w", err)
		}
	}

	// Log the inbound message
	status := "PROGRESS_UPDATE"
	if percent == 100 {
		status = "COMPLETED"
	}
	_ = p.logInboundMessage(ctx, &contact.ID, &taskID, msg, status)

	return nil
}

// handleProgressUpdateTransactional processes a percentage update with transactional idempotency.
// P0 Fix: Ensures atomicity between deduplication and side-effects.
// See PRODUCTION_PLAN.md Phase 49 Retrofit (Operation Ironclad Task 1)
//
// Transaction Flow:
//   - Step A (Lock): INSERT communication_log with external_id FIRST
//   - Step B (Guard): If duplicate key conflict, return nil (already processed)
//   - Step C (Execute): Run side-effects (UpdateTaskProgress, RecalculateSchedule)
//   - Step D (Commit): If any step fails, entire transaction rolls back
func (p *InboundProcessor) handleProgressUpdateTransactional(
	ctx context.Context,
	taskID, projectID uuid.UUID,
	contact *types.Contact,
	percent int,
	msg InboundMessage,
) error {
	// Step A: Begin transaction and attempt to insert communication_log as lock
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx) // No-op if already committed
		}
	}()

	// Determine status for logging
	status := "PROGRESS_UPDATE"
	if percent == 100 {
		status = "COMPLETED"
	}

	// Step A: Insert communication_log FIRST as idempotency lock
	// This uses the external_id unique constraint to prevent duplicate processing
	logID := uuid.New()
	var relatedType *string
	t := "project_task"
	relatedType = &t

	insertQuery := `
		INSERT INTO communication_logs (
			id, project_id, contact_id, direction, content, channel, timestamp,
			related_entity_id, related_entity_type, external_id
		)
		VALUES ($1, $2, $3, 'Inbound', $4, $5, $6, $7, $8, $9)
	`

	// Get project_id from task
	var taskProjectID *uuid.UUID
	pid, pidErr := p.getTaskProjectID(ctx, taskID)
	if pidErr == nil {
		taskProjectID = &pid
	}

	_, err = tx.Exec(ctx, insertQuery,
		logID,
		taskProjectID,
		&contact.ID,
		fmt.Sprintf("[%s] %s", status, msg.Body),
		msg.Channel,
		p.clock.Now(),
		&taskID,
		relatedType,
		nilIfEmpty(msg.ExternalID),
	)

	// Step B: Guard - Check for duplicate key violation
	if err != nil {
		if isDuplicateKeyError(err) {
			p.log.Info("duplicate message detected via transaction lock, skipping side-effects",
				"external_id", msg.ExternalID,
				"task_id", taskID,
			)
			return nil // Already processed - idempotency preserved
		}
		return fmt.Errorf("insert communication log (lock): %w", err)
	}

	// Step C: Execute side-effects now that we hold the lock
	if err := p.schedule.UpdateTaskProgress(ctx, taskID, percent); err != nil {
		// Transaction will rollback, log entry will be removed, allowing clean retry
		return fmt.Errorf("update task progress: %w", err)
	}

	p.log.Info("task progress updated via inbound message (transactional)",
		"task_id", taskID,
		"percent", percent,
		"sender", msg.Sender,
		"channel", msg.Channel,
	)

	// State Machine - 100% triggers CPM recalculation
	if percent == 100 {
		p.log.Info("task marked complete, triggering schedule recalculation",
			"task_id", taskID,
			"project_id", projectID,
		)

		orgID, err := p.getProjectOrgID(ctx, projectID)
		if err != nil {
			return fmt.Errorf("get project org_id: %w", err)
		}

		if err := p.schedule.RecalculateSchedule(ctx, projectID, orgID); err != nil {
			return fmt.Errorf("recalculate schedule: %w", err)
		}
	}

	// Step D: Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	tx = nil // Prevent deferred rollback

	return nil
}

// isDuplicateKeyError checks if the error is a PostgreSQL unique constraint violation.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL error code 23505 = unique_violation
	errStr := err.Error()
	return strings.Contains(errStr, "23505") || strings.Contains(errStr, "duplicate key")
}

// handleVisionVerification triggers AI vision analysis for attached images.
// See PRODUCTION_PLAN.md Step 48 (Vision Trigger)
func (p *InboundProcessor) handleVisionVerification(
	ctx context.Context,
	taskID uuid.UUID,
	contact *types.Contact,
	msg InboundMessage,
) error {
	if p.vision == nil {
		p.log.Warn("vision service not configured, skipping verification",
			"task_id", taskID,
			"image_count", len(msg.ImageURLs),
		)
		return nil
	}

	// Get task description for vision prompt
	taskDescription, err := p.getTaskDescription(ctx, taskID)
	if err != nil {
		return fmt.Errorf("get task description: %w", err)
	}

	// Process first image (MVP - single image support)
	imageURL := msg.ImageURLs[0]
	isVerified, confidence, err := p.vision.VerifyTask(ctx, imageURL, taskDescription)
	if err != nil {
		return fmt.Errorf("vision verify: %w", err)
	}

	p.log.Info("vision verification complete",
		"task_id", taskID,
		"is_verified", isVerified,
		"confidence", confidence,
		"image_url", imageURL,
	)

	// Log the inbound message with vision result
	status := "VISION_PENDING"
	if isVerified && confidence >= 0.8 {
		status = "VISION_VERIFIED"
	}
	_ = p.logInboundMessage(ctx, &contact.ID, &taskID, msg, status)

	return nil
}

// --- Helpers ---

func (p *InboundProcessor) checkExternalIDExists(ctx context.Context, externalID string) (bool, error) {
	query := `SELECT 1 FROM communication_logs WHERE external_id = $1 LIMIT 1`
	var exists int
	err := p.db.QueryRow(ctx, query, externalID).Scan(&exists)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (p *InboundProcessor) findRecentOutboundContext(ctx context.Context, contactID uuid.UUID, window time.Duration) (*uuid.UUID, *uuid.UUID, error) {
	// Query for most recent outbound message to this contact with task context
	query := `
		SELECT related_entity_id, project_id FROM communication_logs
		WHERE contact_id = $1
		  AND direction = 'Outbound'
		  AND related_entity_type = 'project_task'
		  AND related_entity_id IS NOT NULL
		  AND timestamp > $3 - ($2 || ' hours')::interval
		ORDER BY timestamp DESC
		LIMIT 1
	`
	hours := int(window.Hours())
	var taskID, projectID uuid.UUID
	err := p.db.QueryRow(ctx, query, contactID, hours, p.clock.Now()).Scan(&taskID, &projectID)
	if err != nil {
		return nil, nil, err
	}
	return &taskID, &projectID, nil
}

func (p *InboundProcessor) logInboundMessage(ctx context.Context, contactID, taskID *uuid.UUID, msg InboundMessage, status string) error {
	query := `
		INSERT INTO communication_logs (
			id, project_id, contact_id, direction, content, channel, timestamp,
			related_entity_id, related_entity_type, external_id
		)
		VALUES ($1, $2, $3, 'Inbound', $4, $5, $6, $7, $8, $9)
		ON CONFLICT (external_id) WHERE external_id IS NOT NULL DO NOTHING
	`

	// Get project_id from task if available
	var projectID *uuid.UUID
	if taskID != nil {
		pid, err := p.getTaskProjectID(ctx, *taskID)
		if err == nil {
			projectID = &pid
		}
	}

	var relatedType *string
	if taskID != nil {
		t := "project_task"
		relatedType = &t
	}

	_, err := p.db.Exec(ctx, query,
		uuid.New(),
		projectID,
		contactID,
		fmt.Sprintf("[%s] %s", status, msg.Body),
		msg.Channel,
		p.clock.Now(),
		taskID,
		relatedType,
		nilIfEmpty(msg.ExternalID),
	)
	return err
}

func (p *InboundProcessor) getProjectOrgID(ctx context.Context, projectID uuid.UUID) (uuid.UUID, error) {
	query := `SELECT org_id FROM projects WHERE id = $1`
	var orgID uuid.UUID
	err := p.db.QueryRow(ctx, query, projectID).Scan(&orgID)
	return orgID, err
}

func (p *InboundProcessor) getTaskProjectID(ctx context.Context, taskID uuid.UUID) (uuid.UUID, error) {
	query := `SELECT project_id FROM project_tasks WHERE id = $1`
	var projectID uuid.UUID
	err := p.db.QueryRow(ctx, query, taskID).Scan(&projectID)
	return projectID, err
}

func (p *InboundProcessor) getTaskDescription(ctx context.Context, taskID uuid.UUID) (string, error) {
	query := `SELECT name FROM project_tasks WHERE id = $1`
	var name string
	err := p.db.QueryRow(ctx, query, taskID).Scan(&name)
	return name, err
}

func (p *InboundProcessor) createRiskFlag(ctx context.Context, taskID uuid.UUID, message string) error {
	query := `
		INSERT INTO review_flags (id, entity_type, entity_id, reason, status, created_at)
		VALUES ($1, 'project_task', $2, $3, 'Pending', $4)
	`
	reason := fmt.Sprintf("Subcontractor indicated delay/issue: %s", truncateString(message, 200))
	_, err := p.db.Exec(ctx, query, uuid.New(), taskID, reason, p.clock.Now())
	return err
}

func (p *InboundProcessor) parsePercentage(text string) (int, bool) {
	// Match exact percentage format: "50%", "100%"
	matches := inboundPercentageRe.FindStringSubmatch(text)
	if len(matches) > 1 {
		if num, err := strconv.Atoi(matches[1]); err == nil && num >= 0 && num <= 100 {
			return num, true
		}
	}

	// Handle completion keywords
	if text == "done" || text == "complete" || text == "finished" {
		return 100, true
	}

	return 0, false
}

func (p *InboundProcessor) isConfirmation(text string) bool {
	confirmations := []string{"yes", "confirmed", "on site", "on my way", "arriving", "here"}
	for _, c := range confirmations {
		if strings.Contains(text, c) {
			return true
		}
	}
	return false
}

func (p *InboundProcessor) isDelayIndicator(text string) bool {
	indicators := []string{"delay", "delayed", "late", "can't", "cannot", "problem", "issue", "stuck", "waiting"}
	for _, ind := range indicators {
		if strings.Contains(text, ind) {
			return true
		}
	}
	return false
}

// ExtractImageURLs extracts image URLs from message body.
// Exported for use by webhook handler.
func ExtractImageURLs(body string) []string {
	matches := inboundImageURLRe.FindAllString(body, -1)
	return matches
}

// VerifySignature validates webhook signature using HMAC-SHA256.
// See PRODUCTION_PLAN.md Step 48 (Security Amendment)
func VerifySignature(body []byte, signature, secret string) bool {
	if secret == "" {
		return false // Fail closed if no secret configured
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

// --- Feed Card Writers ---

// writeInboundConfirmationCard creates a sub_confirmation card when a sub confirms.
func (p *InboundProcessor) writeInboundConfirmationCard(ctx context.Context, taskID, projectID uuid.UUID, contact *types.Contact, body string) {
	orgID, err := p.getProjectOrgID(ctx, projectID)
	if err != nil {
		p.log.Error("failed to get org_id for confirmation card", "project_id", projectID, "error", err)
		return
	}

	taskName, _ := p.getTaskDescription(ctx, taskID)
	agentSource := "InboundProcessor"
	card := &models.FeedCard{
		OrgID:       orgID,
		ProjectID:   projectID,
		CardType:    models.FeedCardSubConfirmation,
		Priority:    models.FeedCardPriorityLow,
		Headline:    fmt.Sprintf("%s confirmed for %s", contact.Name, taskName),
		Body:        fmt.Sprintf("%s (%s) confirmed: \"%s\"", contact.Name, contact.Company, truncateString(body, 100)),
		Horizon:     models.FeedCardHorizonToday,
		AgentSource: &agentSource,
		TaskID:      &taskID,
		Actions: []models.FeedCardAction{
			{ID: "dismiss", Label: "Got It", Style: "primary"},
		},
	}

	if err := p.feedWriter.WriteCard(ctx, card); err != nil {
		p.log.Error("failed to write sub_confirmation feed card", "task_id", taskID, "error", err)
	}
}

// writeInboundDelayCard creates a sub_unconfirmed card when a sub reports a delay.
func (p *InboundProcessor) writeInboundDelayCard(ctx context.Context, taskID, projectID uuid.UUID, contact *types.Contact, body string) {
	orgID, err := p.getProjectOrgID(ctx, projectID)
	if err != nil {
		p.log.Error("failed to get org_id for delay card", "project_id", projectID, "error", err)
		return
	}

	taskName, _ := p.getTaskDescription(ctx, taskID)
	agentSource := "InboundProcessor"
	consequence := "Subcontractor indicated a potential delay. Review impact on schedule."
	card := &models.FeedCard{
		OrgID:       orgID,
		ProjectID:   projectID,
		CardType:    models.FeedCardSubUnconfirmed,
		Priority:    models.FeedCardPriorityUrgent,
		Headline:    fmt.Sprintf("Delay reported by %s for %s", contact.Name, taskName),
		Body:        fmt.Sprintf("%s (%s) reported: \"%s\"", contact.Name, contact.Company, truncateString(body, 100)),
		Consequence: &consequence,
		Horizon:     models.FeedCardHorizonToday,
		AgentSource: &agentSource,
		TaskID:      &taskID,
		Actions: []models.FeedCardAction{
			{ID: "view_schedule", Label: "View Schedule Impact", Style: "primary"},
			{ID: "call_sub", Label: "Call Sub", Style: "secondary"},
			{ID: "dismiss", Label: "Dismiss", Style: "secondary"},
		},
	}

	if err := p.feedWriter.WriteCard(ctx, card); err != nil {
		p.log.Error("failed to write sub delay feed card", "task_id", taskID, "error", err)
	}
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
