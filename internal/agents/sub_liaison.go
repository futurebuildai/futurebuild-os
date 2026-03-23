// Package agents provides AI-powered business logic components for FutureBuild.
package agents

import (
	"context"
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
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Package-Level Compiled Regexes ---
// ENGINEERING STANDARD: Compile once at startup, not per-call.
// See Code Review WARN #1 (Hot Path Optimization)
var (
	percentageRe = regexp.MustCompile(`(\d+)\s*%|(\d+)\s*percent`)
	imageURLRe   = regexp.MustCompile(`https?://[^\s]+\.(jpg|jpeg|png|gif|webp)`)
)

// SubLiaisonAgent handles outbound subcontractor coordination and inbound status updates.
// This is the "Virtual Superintendent" for trade communication.
// See PRODUCTION_PLAN.md Step 47
// Refactored for deterministic simulation: PRODUCTION_PLAN.md Step 49
type SubLiaisonAgent struct {
	db           *pgxpool.Pool
	directory    DirectoryService
	notifier     NotificationService
	clock        clock.Clock
	feedWriter   FeedWriter   // V2: writes sub_confirmation/sub_unconfirmed cards
	claudeRunner *AgentRunner // Claude reasoning for nuanced SMS understanding
}

// DirectoryService defines contact lookup operations.
// See PRODUCTION_PLAN.md Step 47 (Interface for testability)
type DirectoryService interface {
	GetContactForPhase(ctx context.Context, projectID, orgID uuid.UUID, phaseCode string) (*types.Contact, error)
}

// NotificationService defines outbound communication operations.
// See PRODUCTION_PLAN.md Step 47 (Interface for testability)
type NotificationService interface {
	SendSMS(contactID string, message string) error
	SendEmail(to string, subject string, body string) error
}

// NewSubLiaisonAgent creates a new agent with injected dependencies.
// Clock is required for deterministic time simulation (Step 49).
func NewSubLiaisonAgent(db *pgxpool.Pool, directory DirectoryService, notifier NotificationService, clk clock.Clock) *SubLiaisonAgent {
	return &SubLiaisonAgent{
		db:        db,
		directory: directory,
		notifier:  notifier,
		clock:     clk,
	}
}

// WithClaudeRunner sets the AgentRunner for Claude-powered message understanding.
func (a *SubLiaisonAgent) WithClaudeRunner(runner *AgentRunner) *SubLiaisonAgent {
	a.claudeRunner = runner
	return a
}

// WithFeedWriter sets the feed writer for V2 portfolio feed card generation.
// If not set, no feed cards are written (backward compatible).
func (a *SubLiaisonAgent) WithFeedWriter(fw FeedWriter) *SubLiaisonAgent {
	a.feedWriter = fw
	return a
}

// ScanAndNotify finds tasks starting within 72h and sends confirmation requests.
// See PRODUCTION_PLAN.md Step 49 Amendment 2
func (a *SubLiaisonAgent) ScanAndNotify(ctx context.Context) error {
	windowStart := a.clock.Now()
	windowEnd := a.clock.Now().Add(72 * time.Hour)

	// Query tasks where early_start is between windowStart and windowEnd
	// and status is not 'Completed'
	query := `
		SELECT id FROM project_tasks
		WHERE early_start >= $1
		  AND early_start <= $2
		  AND status != 'Completed'
	`
	rows, err := a.db.Query(ctx, query, windowStart, windowEnd)
	if err != nil {
		return fmt.Errorf("failed to query upcoming tasks: %w", err)
	}
	defer rows.Close()

	var taskIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			slog.Error("Failed to scan task ID", "error", err)
			continue
		}
		taskIDs = append(taskIDs, id)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows iteration error: %w", err)
	}

	// Iterate and send confirmations, logging errors but continuing
	var errs []error
	for _, taskID := range taskIDs {
		if err := a.ConfirmArrival(ctx, taskID); err != nil {
			slog.Error("Failed to confirm arrival for task", "task_id", taskID, "error", err)
			errs = append(errs, err)
		}
	}

	slog.Info("ScanAndNotify completed", "tasks_found", len(taskIDs), "errors", len(errs))
	return nil
}

// CommunicationLogType defines the type of communication log entry.
// See DATA_SPINE_SPEC.md Section 5.1
type CommunicationLogType string

const (
	CommLogConfirmationRequest CommunicationLogType = "CONFIRMATION_REQUEST"
	CommLogStatusRequest       CommunicationLogType = "STATUS_REQUEST"
	CommLogInboundReply        CommunicationLogType = "INBOUND_REPLY"
)

// SendStatus defines the delivery status for the Transactional Outbox pattern.
// See P0 Fix: Non-Atomic Side Effects - ensures at-most-once delivery.
type SendStatus string

const (
	SendStatusPending SendStatus = "PENDING"
	SendStatusSent    SendStatus = "SENT"
	SendStatusFailed  SendStatus = "FAILED"
)

// ConfirmArrival sends a confirmation request to the subcontractor assigned to a task's phase.
// See PRODUCTION_PLAN.md Step 47
//
// TRANSACTIONAL OUTBOX PATTERN (P0 Fix: At-Most-Once Delivery)
// Logic Flow:
// 1. Fetch task to get project_id, wbs_code, early_start
// 2. Extract phase code (e.g., "9.1" -> "9")
// 3. Call DirectoryService.GetContactForPhase
// 4. Guard: If no contact, log WARN and return nil (expected state)
// 5. Idempotency: Check communication_logs for recent SENT request (24h)
// 6. OUTBOX: Insert PENDING record (durable before side effect)
// 7. Send notification based on contact preference
// 8. Update record to SENT or FAILED
func (a *SubLiaisonAgent) ConfirmArrival(ctx context.Context, taskID uuid.UUID) error {
	// Step 1: Fetch task details
	// See DATA_SPINE_SPEC.md Section 3.3
	task, err := a.getTaskDetails(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to fetch task %s: %w", taskID, err)
	}

	// Step 2: Extract phase code from WBS code
	// Convention: "9.1.2" -> "9" (major phase group)
	phaseCode := extractPhaseCode(task.WBSCode)

	// Step 3: Resolve contact for this phase
	contact, err := a.directory.GetContactForPhase(ctx, task.ProjectID, task.OrgID, phaseCode)
	if err != nil {
		// Guard Clause: Missing contact is expected (not all phases have assigned subs)
		// Log WARN per spec, do not fail the job
		slog.Warn("No contact for phase",
			"phase_code", phaseCode,
			"project_id", task.ProjectID,
			"task_id", taskID,
			"error", err,
		)

		// V2 Feed: Write setup_contacts card prompting builder to assign a contact
		// See FRONTEND_V2_SPEC.md §10.3.C
		if a.feedWriter != nil {
			a.writeSetupContactsCard(ctx, task, phaseCode)
		}
		return nil
	}

	// Step 4: Idempotency check (24h dampening)
	// Only considers SENT records - FAILED can be retried, PENDING handled by outbox processor
	recentExists, err := a.hasRecentCommunication(ctx, contact.ID, taskID, 24*time.Hour)
	if err != nil {
		slog.Error("Failed to check communication log", "error", err)
		return nil // Non-fatal, skip notification
	}
	if recentExists {
		slog.Info("Skipping duplicate confirmation request",
			"contact_id", contact.ID,
			"task_id", taskID,
		)
		return nil
	}

	// Step 5: Build message
	message := fmt.Sprintf(
		"FutureBuild: Please confirm arrival for '%s' scheduled %s. Reply YES to confirm or provide status update.",
		task.Name,
		formatDate(task.EarlyStart),
	)

	// Step 6: TRANSACTIONAL OUTBOX - Insert PENDING record BEFORE external call
	// This ensures durability: if we crash after send, the record exists for retry detection
	logID, err := a.logCommunicationWithStatus(ctx, task.ProjectID, contact.ID, &taskID, CommLogConfirmationRequest, message, SendStatusPending)
	if err != nil {
		// Failed to create log - don't send SMS (preserves at-most-once guarantee)
		return fmt.Errorf("failed to create pending log: %w", err)
	}

	// Step 7: Send notification (record is now durable)
	sendErr := a.sendNotification(contact, message)

	// Step 8: Update status based on send result
	finalStatus := SendStatusSent
	if sendErr != nil {
		finalStatus = SendStatusFailed
		slog.Error("Failed to send confirmation notification",
			"contact_id", contact.ID,
			"log_id", logID,
			"error", sendErr,
		)
	}
	if err := a.updateCommunicationStatus(ctx, logID, finalStatus); err != nil {
		slog.Error("Failed to update log status", "log_id", logID, "status", finalStatus, "error", err)
	}

	if sendErr == nil {
		slog.Info("Confirmation request sent",
			"task_id", taskID,
			"contact_id", contact.ID,
			"phase_code", phaseCode,
			"log_id", logID,
		)

		// V2 Feed: Write sub_unconfirmed card (awaiting response)
		if a.feedWriter != nil {
			a.writeSubUnconfirmedCard(ctx, task, contact)
		}
	}

	return nil
}

// HandleInboundMessage processes an incoming message from a subcontractor.
// See PRODUCTION_PLAN.md Step 47 (Amendment #1: Context Binding)
//
// Logic Flow:
// 1. Match sender (phone/email) to a Contact
// 2. Query communication_logs for most recent outbound request (48h) to bind context
// 3. Parse body for: percentage -> update progress, "delay"/"no" -> create flag
func (a *SubLiaisonAgent) HandleInboundMessage(ctx context.Context, sender, body string) error {
	// Step 1: Resolve sender to contact
	contact, err := a.findContactBySender(ctx, sender)
	if err != nil {
		slog.Warn("Unknown sender", "sender", sender, "error", err)
		return nil // Unknown sender, ignore
	}

	// Step 2: Context Binding (Amendment #1)
	// Find the most recent outbound request to this contact (48h window)
	taskID, err := a.findRecentOutboundTask(ctx, contact.ID, 48*time.Hour)
	if err != nil {
		slog.Warn("No recent outbound context for sender",
			"contact_id", contact.ID,
			"sender", sender,
		)
		return nil // No context to bind, ignore
	}

	// Step 3: Parse body and take action
	normalizedBody := strings.ToLower(strings.TrimSpace(body))

	// Check for percentage update (e.g., "50%", "75 %", "done")
	if percent, ok := parsePercentage(normalizedBody); ok {
		if err := a.updateTaskProgress(ctx, *taskID, percent); err != nil {
			slog.Error("Failed to update task progress", "task_id", taskID, "error", err)
		} else {
			slog.Info("Task progress updated via SMS",
				"task_id", taskID,
				"percent", percent,
				"sender", sender,
			)
		}
	}

	// Check for delay/risk indicators
	if containsDelayIndicator(normalizedBody) {
		if err := a.createRiskFlag(ctx, *taskID, body); err != nil {
			slog.Error("Failed to create risk flag", "task_id", taskID, "error", err)
		} else {
			slog.Info("Risk flag created from inbound message",
				"task_id", taskID,
				"sender", sender,
			)
		}
	}

	// Check for image URL (stub for VisionService)
	if imageURL := extractImageURL(normalizedBody); imageURL != "" {
		// TODO: Call VisionService.Analyze when Vision integration is ready
		slog.Info("Image URL received (VisionService stub)",
			"task_id", taskID,
			"image_url", imageURL,
		)
	}

	// Log inbound communication
	if err := a.logCommunication(ctx, uuid.Nil, contact.ID, taskID, CommLogInboundReply, body); err != nil {
		slog.Error("Failed to log inbound communication", "error", err)
	}

	// V2 Feed: Write confirmation or delay card based on response
	if a.feedWriter != nil && taskID != nil {
		task, taskErr := a.getTaskDetails(ctx, *taskID)
		if taskErr == nil {
			isConfirmation := strings.Contains(normalizedBody, "yes") ||
				strings.Contains(normalizedBody, "confirm") ||
				strings.Contains(normalizedBody, "done") ||
				strings.Contains(normalizedBody, "complete")
			isDelay := containsDelayIndicator(normalizedBody)

			if isConfirmation {
				a.writeSubConfirmationCard(ctx, task, contact, body)
			} else if isDelay {
				a.writeSubDelayCard(ctx, task, contact, body)
			}
		}
	}

	return nil
}

// --- Feed Card Writers ---

// writeSubUnconfirmedCard creates a sub_unconfirmed card when a confirmation request is sent.
func (a *SubLiaisonAgent) writeSubUnconfirmedCard(ctx context.Context, task *taskDetails, contact *types.Contact) {
	agentSource := "SubLiaisonAgent"
	card := &models.FeedCard{
		OrgID:    task.OrgID,
		ProjectID: task.ProjectID,
		CardType: models.FeedCardSubUnconfirmed,
		Priority: models.FeedCardPriorityNormal,
		Headline: fmt.Sprintf("Awaiting confirmation from %s for %s", contact.Name, task.Name),
		Body:     fmt.Sprintf("Confirmation request sent to %s (%s). Scheduled start: %s.", contact.Name, contact.Company, formatDate(task.EarlyStart)),
		Horizon:  models.FeedCardHorizonThisWeek,
		Deadline: task.EarlyStart,
		AgentSource: &agentSource,
		TaskID:   &task.ID,
		Actions: []models.FeedCardAction{
			{ID: "resend", Label: "Resend Request", Style: "primary"},
			{ID: "call_sub", Label: "Call Sub", Style: "secondary"},
			{ID: "dismiss", Label: "Dismiss", Style: "secondary"},
		},
	}

	if task.EarlyStart != nil {
		daysUntil := int(task.EarlyStart.Sub(a.clock.Now()).Hours() / 24)
		if daysUntil <= 1 {
			card.Priority = models.FeedCardPriorityUrgent
			card.Horizon = models.FeedCardHorizonToday
			c := "Task starts soon with no sub confirmation."
			card.Consequence = &c
		}
	}

	if err := a.feedWriter.WriteCard(ctx, card); err != nil {
		slog.Error("failed to write sub_unconfirmed feed card", "task_id", task.ID, "error", err)
	}
}

// writeSubConfirmationCard creates a sub_confirmation card when a sub confirms arrival.
func (a *SubLiaisonAgent) writeSubConfirmationCard(ctx context.Context, task *taskDetails, contact *types.Contact, body string) {
	agentSource := "SubLiaisonAgent"
	card := &models.FeedCard{
		OrgID:    task.OrgID,
		ProjectID: task.ProjectID,
		CardType: models.FeedCardSubConfirmation,
		Priority: models.FeedCardPriorityLow,
		Headline: fmt.Sprintf("%s confirmed for %s", contact.Name, task.Name),
		Body:     fmt.Sprintf("%s (%s) confirmed: \"%s\"", contact.Name, contact.Company, truncateString(body, 100)),
		Horizon:  models.FeedCardHorizonToday,
		AgentSource: &agentSource,
		TaskID:   &task.ID,
		Actions: []models.FeedCardAction{
			{ID: "dismiss", Label: "Got It", Style: "primary"},
		},
	}

	if err := a.feedWriter.WriteCard(ctx, card); err != nil {
		slog.Error("failed to write sub_confirmation feed card", "task_id", task.ID, "error", err)
	}
}

// writeSubDelayCard creates a sub_unconfirmed card with elevated priority when a sub indicates a delay.
func (a *SubLiaisonAgent) writeSubDelayCard(ctx context.Context, task *taskDetails, contact *types.Contact, body string) {
	agentSource := "SubLiaisonAgent"
	c := "Subcontractor indicated a potential delay. Review impact on schedule."
	card := &models.FeedCard{
		OrgID:       task.OrgID,
		ProjectID:   task.ProjectID,
		CardType:    models.FeedCardSubUnconfirmed,
		Priority:    models.FeedCardPriorityUrgent,
		Headline:    fmt.Sprintf("Delay reported by %s for %s", contact.Name, task.Name),
		Body:        fmt.Sprintf("%s (%s) reported: \"%s\"", contact.Name, contact.Company, truncateString(body, 100)),
		Consequence: &c,
		Horizon:     models.FeedCardHorizonToday,
		AgentSource: &agentSource,
		TaskID:      &task.ID,
		Actions: []models.FeedCardAction{
			{ID: "view_schedule", Label: "View Schedule Impact", Style: "primary"},
			{ID: "call_sub", Label: "Call Sub", Style: "secondary"},
			{ID: "dismiss", Label: "Dismiss", Style: "secondary"},
		},
	}

	if err := a.feedWriter.WriteCard(ctx, card); err != nil {
		slog.Error("failed to write sub delay feed card", "task_id", task.ID, "error", err)
	}
}

// writeSetupContactsCard creates a setup_contacts card when a phase has no assigned contact
// and a task in that phase is approaching. The card includes an inline form for quick assignment.
// See FRONTEND_V2_SPEC.md §10.3.C
func (a *SubLiaisonAgent) writeSetupContactsCard(ctx context.Context, task *taskDetails, phaseCode string) {
	tradeName := phaseTradeNames[phaseCode]
	if tradeName == "" {
		tradeName = "Phase " + phaseCode
	}

	headline := fmt.Sprintf("No %s contact assigned", strings.ToLower(tradeName))
	body := fmt.Sprintf("%s starts %s. I need a contact to send a start confirmation.", task.Name, formatDate(task.EarlyStart))

	agentSource := "SubLiaisonAgent"
	card := &models.FeedCard{
		OrgID:       task.OrgID,
		ProjectID:   task.ProjectID,
		CardType:    models.FeedCardSetupContacts,
		Priority:    models.FeedCardPriorityUrgent,
		Headline:    headline,
		Body:        body,
		Horizon:     models.FeedCardHorizonToday,
		Deadline:    task.EarlyStart,
		AgentSource: &agentSource,
		TaskID:      &task.ID,
		Actions: []models.FeedCardAction{
			{ID: "assign_contact", Label: "Save & Assign", Style: "primary"},
			{ID: "dismiss", Label: "Skip this phase", Style: "secondary"},
		},
	}

	if err := a.feedWriter.WriteCard(ctx, card); err != nil {
		slog.Error("failed to write setup_contacts feed card", "task_id", task.ID, "phase_code", phaseCode, "error", err)
	}
}

// phaseTradeNames maps WBS phase codes to human-readable trade names.
var phaseTradeNames = map[string]string{
	"5":  "Permit & Site Prep",
	"6":  "Foundation",
	"7":  "Framing",
	"8":  "Roofing & Exterior",
	"9":  "Rough-Ins (MEP)",
	"10": "Insulation & Drywall",
	"11": "Finishes",
	"12": "Final Inspections",
	"13": "Punch List & Closeout",
}

// --- Private Helper Methods ---

// taskDetails holds the minimal task info needed for liaison operations.
type taskDetails struct {
	ID         uuid.UUID
	ProjectID  uuid.UUID
	OrgID      uuid.UUID
	WBSCode    string
	Name       string
	EarlyStart *time.Time
}

func (a *SubLiaisonAgent) getTaskDetails(ctx context.Context, taskID uuid.UUID) (*taskDetails, error) {
	query := `
		SELECT pt.id, pt.project_id, p.org_id, pt.wbs_code, pt.name, pt.early_start
		FROM project_tasks pt
		JOIN projects p ON pt.project_id = p.id
		WHERE pt.id = $1
	`
	var t taskDetails
	err := a.db.QueryRow(ctx, query, taskID).Scan(
		&t.ID, &t.ProjectID, &t.OrgID, &t.WBSCode, &t.Name, &t.EarlyStart,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// extractPhaseCode extracts the major phase from a WBS code.
// Uses types.WBSCode value object for robust format handling.
// Technical Debt Remediation (P2): Supports various formats beyond simple X.Y.Z
func extractPhaseCode(wbsCode string) string {
	parsed, err := types.ParseWBSCode(wbsCode)
	if err != nil {
		slog.Warn("Invalid WBS code format, using raw value", "wbs_code", wbsCode, "error", err)
		return wbsCode
	}
	return parsed.GetMajorPhase()
}

func (a *SubLiaisonAgent) hasRecentCommunication(ctx context.Context, contactID, taskID uuid.UUID, window time.Duration) (bool, error) {
	// Check communication_logs for recent SENT outbound to this contact for this task
	// Uses 'related_entity_id' column added in migration 000046
	// Only considers SENT records - FAILED can be retried, PENDING handled by outbox processor
	// ENGINEERING STANDARD: Use explicit hours for PostgreSQL interval (Code Review WARN #2)
	// See PRODUCTION_PLAN.md Step 49: Uses injected clock for deterministic simulation.
	query := `
		SELECT COUNT(*) FROM communication_logs
		WHERE contact_id = $1
		  AND related_entity_id = $2
		  AND timestamp > ($4::timestamptz - ($3 || ' hours')::interval)
		  AND direction = 'Outbound'
		  AND send_status = 'SENT'
	`
	hours := fmt.Sprintf("%d", int(window.Hours()))
	var count int
	err := a.db.QueryRow(ctx, query, contactID, taskID, hours, a.clock.Now()).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (a *SubLiaisonAgent) sendNotification(contact *types.Contact, message string) error {
	switch contact.ContactPreference {
	case types.ContactPreferenceSMS:
		return a.notifier.SendSMS(contact.ID.String(), message)
	case types.ContactPreferenceEmail:
		return a.notifier.SendEmail(contact.Email, "FutureBuild: Arrival Confirmation", message)
	case types.ContactPreferenceBoth:
		// Send both, log errors but don't fail
		if err := a.notifier.SendSMS(contact.ID.String(), message); err != nil {
			slog.Warn("SMS send failed, trying email", "error", err)
		}
		return a.notifier.SendEmail(contact.Email, "FutureBuild: Arrival Confirmation", message)
	default:
		// Default to SMS for Subcontractor preference
		return a.notifier.SendSMS(contact.ID.String(), message)
	}
}

func (a *SubLiaisonAgent) logCommunication(ctx context.Context, projectID, contactID uuid.UUID, taskID *uuid.UUID, logType CommunicationLogType, content string) error {
	// Wrapper for backwards compatibility - defaults to SENT status
	_, err := a.logCommunicationWithStatus(ctx, projectID, contactID, taskID, logType, content, SendStatusSent)
	return err
}

// logCommunicationWithStatus inserts a communication log with explicit status and returns the log ID.
// This is the core of the Transactional Outbox pattern - insert PENDING before external call.
func (a *SubLiaisonAgent) logCommunicationWithStatus(ctx context.Context, projectID, contactID uuid.UUID, taskID *uuid.UUID, logType CommunicationLogType, content string, status SendStatus) (uuid.UUID, error) {
	// See PRODUCTION_PLAN.md Step 49: Uses injected clock for deterministic simulation.
	query := `
		INSERT INTO communication_logs (
			id, project_id, contact_id, direction, content, channel, timestamp,
			related_entity_id, related_entity_type, send_status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`
	direction := "Outbound"
	if logType == CommLogInboundReply {
		direction = "Inbound"
	}

	var relatedID *uuid.UUID
	if taskID != nil && *taskID != uuid.Nil {
		relatedID = taskID
	}

	// Get project_id from task if not provided
	if projectID == uuid.Nil && relatedID != nil {
		task, err := a.getTaskDetails(ctx, *relatedID)
		if err == nil {
			projectID = task.ProjectID
		}
	}

	logID := uuid.New()
	err := a.db.QueryRow(ctx, query,
		logID, projectID, contactID, direction, content, "SMS", a.clock.Now(),
		relatedID, "project_task", string(status),
	).Scan(&logID)
	if err != nil {
		return uuid.Nil, err
	}
	return logID, nil
}

// updateCommunicationStatus updates the send_status of a communication log by ID.
// Used to transition PENDING -> SENT or PENDING -> FAILED after external call.
func (a *SubLiaisonAgent) updateCommunicationStatus(ctx context.Context, logID uuid.UUID, status SendStatus) error {
	query := `UPDATE communication_logs SET send_status = $1 WHERE id = $2`
	_, err := a.db.Exec(ctx, query, string(status), logID)
	return err
}

func (a *SubLiaisonAgent) findContactBySender(ctx context.Context, sender string) (*types.Contact, error) {
	// Normalize phone number (strip non-digits) or use as email
	query := `
		SELECT id, name, company, COALESCE(phone, ''), COALESCE(email, ''), role, contact_preference
		FROM contacts
		WHERE phone = $1 OR email = $1
		LIMIT 1
	`
	var contact types.Contact
	var role, preference string
	err := a.db.QueryRow(ctx, query, sender).Scan(
		&contact.ID, &contact.Name, &contact.Company, &contact.Phone, &contact.Email,
		&role, &preference,
	)
	if err != nil {
		return nil, err
	}
	contact.Role = types.UserRole(role)
	contact.ContactPreference = types.ContactPreference(preference)
	return &contact, nil
}

// findRecentOutboundTask finds the task_id from the most recent outbound communication to this contact.
// See PRODUCTION_PLAN.md Step 47 (Amendment #1: Context Binding)
// See PRODUCTION_PLAN.md Step 49: Uses injected clock for deterministic simulation.
func (a *SubLiaisonAgent) findRecentOutboundTask(ctx context.Context, contactID uuid.UUID, window time.Duration) (*uuid.UUID, error) {
	// ENGINEERING STANDARD: Use explicit hours for PostgreSQL interval (Code Review WARN #2)
	query := `
		SELECT related_entity_id FROM communication_logs
		WHERE contact_id = $1
		  AND direction = 'Outbound'
		  AND related_entity_type = 'project_task'
		  AND timestamp > ($3::timestamptz - ($2 || ' hours')::interval)
		ORDER BY timestamp DESC
		LIMIT 1
	`
	hours := fmt.Sprintf("%d", int(window.Hours()))
	var taskID uuid.UUID
	err := a.db.QueryRow(ctx, query, contactID, hours, a.clock.Now()).Scan(&taskID)
	if err != nil {
		return nil, err
	}
	return &taskID, nil
}

func (a *SubLiaisonAgent) updateTaskProgress(ctx context.Context, taskID uuid.UUID, percent int) error {
	// Update percent_complete on project_tasks
	// Note: percent_complete column may not exist in current schema, fallback to task_progress table
	// See PRODUCTION_PLAN.md Step 49: Uses injected clock for deterministic simulation.
	query := `
		INSERT INTO task_progress (id, task_id, reported_by, reported_at, percent_complete, notes)
		VALUES ($1, $2, NULL, $3, $4, 'Updated via SMS')
	`
	_, err := a.db.Exec(ctx, query, uuid.New(), taskID, a.clock.Now(), percent)
	return err
}

func (a *SubLiaisonAgent) createRiskFlag(ctx context.Context, taskID uuid.UUID, originalMessage string) error {
	// Insert into project_flags table (create flag for PM review)
	// Note: project_flags table structure assumed from DATA_SPINE_SPEC.md
	// See PRODUCTION_PLAN.md Step 49: Uses injected clock for deterministic simulation.
	query := `
		INSERT INTO review_flags (id, entity_type, entity_id, reason, status, created_at)
		VALUES ($1, 'project_task', $2, $3, 'Pending', $4)
	`
	reason := fmt.Sprintf("Subcontractor indicated delay/issue: %s", truncateString(originalMessage, 200))
	_, err := a.db.Exec(ctx, query, uuid.New(), taskID, reason, a.clock.Now())
	return err
}

// --- Parsing Helpers ---

// parsePercentage extracts a percentage from text like "50%", "75 percent", "done"
// Uses package-level compiled regex for performance (Code Review WARN #1)
func parsePercentage(text string) (int, bool) {
	// Priority 1: Match patterns like "50%", "50 %", "50percent"
	// This must come BEFORE keyword check to avoid shadowing
	matches := percentageRe.FindStringSubmatch(text)
	if len(matches) > 1 {
		numStr := matches[1]
		if numStr == "" {
			numStr = matches[2]
		}
		if num, err := strconv.Atoi(numStr); err == nil && num >= 0 && num <= 100 {
			return num, true
		}
	}

	// Priority 2: Handle "done" / "complete" / "finished" as 100%
	// Only if no numeric percentage was found
	if strings.Contains(text, "done") || strings.Contains(text, "complete") || strings.Contains(text, "finished") {
		return 100, true
	}

	return 0, false
}

// containsDelayIndicator checks for delay/issue keywords
func containsDelayIndicator(text string) bool {
	indicators := []string{"delay", "delayed", "late", "no", "can't", "cannot", "problem", "issue", "stuck", "waiting"}
	for _, ind := range indicators {
		if strings.Contains(text, ind) {
			return true
		}
	}
	return false
}

// extractImageURL finds an image URL in text (basic pattern)
// Uses package-level compiled regex for performance (Code Review WARN #1)
func extractImageURL(text string) string {
	return imageURLRe.FindString(text)
}

func formatDate(t *time.Time) string {
	if t == nil {
		return "TBD"
	}
	return t.Format("Jan 02, 2006")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
