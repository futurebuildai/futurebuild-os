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
	db        *pgxpool.Pool
	directory DirectoryService
	notifier  NotificationService
	clock     clock.Clock
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

// ConfirmArrival sends a confirmation request to the subcontractor assigned to a task's phase.
// See PRODUCTION_PLAN.md Step 47
//
// Logic Flow:
// 1. Fetch task to get project_id, wbs_code, early_start
// 2. Extract phase code (e.g., "9.1" -> "9")
// 3. Call DirectoryService.GetContactForPhase
// 4. Guard: If no contact, log WARN and return nil (expected state)
// 5. Idempotency: Check communication_logs for recent request (24h)
// 6. Send notification based on contact preference
// 7. Audit: Insert into communication_logs
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
		return nil
	}

	// Step 4: Idempotency check (24h dampening)
	// See User Amendment #4 from Procurement Agent
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

	// Step 5: Build and send notification
	message := fmt.Sprintf(
		"FutureBuild: Please confirm arrival for '%s' scheduled %s. Reply YES to confirm or provide status update.",
		task.Name,
		formatDate(task.EarlyStart),
	)

	if err := a.sendNotification(contact, message); err != nil {
		slog.Error("Failed to send confirmation notification",
			"contact_id", contact.ID,
			"error", err,
		)
		return nil // Non-fatal
	}

	// Step 6: Audit log
	if err := a.logCommunication(ctx, task.ProjectID, contact.ID, &taskID, CommLogConfirmationRequest, message); err != nil {
		slog.Error("Failed to log communication", "error", err)
	}

	slog.Info("Confirmation request sent",
		"task_id", taskID,
		"contact_id", contact.ID,
		"phase_code", phaseCode,
	)

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

	return nil
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
// Convention: "9.1.2" -> "9", "14.3" -> "14"
func extractPhaseCode(wbsCode string) string {
	parts := strings.Split(wbsCode, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return wbsCode
}

func (a *SubLiaisonAgent) hasRecentCommunication(ctx context.Context, contactID, taskID uuid.UUID, window time.Duration) (bool, error) {
	// Check communication_logs for recent outbound to this contact for this task
	// Uses 'related_entity_id' column added in migration 000046
	// ENGINEERING STANDARD: Use explicit hours for PostgreSQL interval (Code Review WARN #2)
	// See PRODUCTION_PLAN.md Step 49: Uses injected clock for deterministic simulation.
	query := `
		SELECT COUNT(*) FROM communication_logs
		WHERE contact_id = $1
		  AND related_entity_id = $2
		  AND timestamp > $4 - ($3 || ' hours')::interval
		  AND direction = 'Outbound'
	`
	hours := int(window.Hours())
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
	// See PRODUCTION_PLAN.md Step 49: Uses injected clock for deterministic simulation.
	query := `
		INSERT INTO communication_logs (
			project_id, contact_id, direction, content, channel, timestamp,
			related_entity_id, related_entity_type
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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

	_, err := a.db.Exec(ctx, query,
		projectID, contactID, direction, content, "SMS", a.clock.Now(),
		relatedID, "project_task",
	)
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
		  AND timestamp > $3 - ($2 || ' hours')::interval
		ORDER BY timestamp DESC
		LIMIT 1
	`
	hours := int(window.Hours())
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
