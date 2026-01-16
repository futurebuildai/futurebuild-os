// Package agents provides AI-powered business logic components for FutureBuild.
package agents

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/colton/futurebuild/pkg/clock"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OutboxSweeper handles orphaned PENDING communication logs.
// This is the recovery process for the Transactional Outbox pattern.
// See Code Review Issue 1A: Rows stuck in PENDING state after crashes.
//
// Logic:
// 1. Query communication_logs for PENDING records older than threshold
// 2. For each record, attempt to re-send notification
// 3. Update status to SENT or FAILED (with retry limit)
type OutboxSweeper struct {
	db        *pgxpool.Pool
	directory DirectoryService
	notifier  NotificationService
	clock     clock.Clock
	config    OutboxSweeperConfig
}

// OutboxSweeperConfig holds configurable parameters for the sweeper.
type OutboxSweeperConfig struct {
	// StaleThreshold is how old a PENDING record must be to be considered orphaned.
	// Default: 5 minutes
	StaleThreshold time.Duration

	// MaxRetries is the maximum number of retry attempts before marking FAILED.
	// Default: 3
	MaxRetries int

	// BatchSize limits the number of records processed per sweep.
	// Default: 100
	BatchSize int
}

// DefaultOutboxSweeperConfig returns sensible defaults for the sweeper.
func DefaultOutboxSweeperConfig() OutboxSweeperConfig {
	return OutboxSweeperConfig{
		StaleThreshold: 5 * time.Minute,
		MaxRetries:     3,
		BatchSize:      100,
	}
}

// OrphanedLog represents a communication log stuck in PENDING state.
type OrphanedLog struct {
	ID        uuid.UUID
	ContactID uuid.UUID
	Content   string
	Channel   string
	Timestamp time.Time
	// RetryCount tracks how many times we've attempted to process this record.
	// Stored as JSONB metadata or a dedicated column; defaults to 0 if not present.
	RetryCount int
}

// NewOutboxSweeper creates a new sweeper with injected dependencies.
func NewOutboxSweeper(
	db *pgxpool.Pool,
	directory DirectoryService,
	notifier NotificationService,
	clk clock.Clock,
	config OutboxSweeperConfig,
) *OutboxSweeper {
	return &OutboxSweeper{
		db:        db,
		directory: directory,
		notifier:  notifier,
		clock:     clk,
		config:    config,
	}
}

// Execute scans for orphaned PENDING logs and attempts recovery.
// This should be called periodically by a cron job or Asynq scheduler.
// See PRODUCTION_PLAN.md: Transactional Outbox Pattern Recovery
func (s *OutboxSweeper) Execute(ctx context.Context) error {
	orphaned, err := s.findOrphanedPending(ctx)
	if err != nil {
		return fmt.Errorf("failed to find orphaned logs: %w", err)
	}

	if len(orphaned) == 0 {
		slog.Debug("OutboxSweeper: No orphaned PENDING logs found")
		return nil
	}

	slog.Info("OutboxSweeper: Processing orphaned logs", "count", len(orphaned))

	var successCount, failCount int
	for _, log := range orphaned {
		if err := s.processOrphanedLog(ctx, log); err != nil {
			slog.Error("OutboxSweeper: Failed to process log",
				"log_id", log.ID,
				"error", err,
			)
			failCount++
		} else {
			successCount++
		}
	}

	slog.Info("OutboxSweeper: Sweep completed",
		"processed", len(orphaned),
		"success", successCount,
		"failed", failCount,
	)

	return nil
}

// findOrphanedPending queries for PENDING logs older than the threshold.
func (s *OutboxSweeper) findOrphanedPending(ctx context.Context) ([]OrphanedLog, error) {
	threshold := s.clock.Now().Add(-s.config.StaleThreshold)

	query := `
		SELECT id, contact_id, content, channel, timestamp, 
		       COALESCE((metadata->>'retry_count')::int, 0) as retry_count
		FROM communication_logs
		WHERE send_status = 'PENDING'
		  AND timestamp < $1
		  AND direction = 'Outbound'
		ORDER BY timestamp ASC
		LIMIT $2
	`

	rows, err := s.db.Query(ctx, query, threshold, s.config.BatchSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []OrphanedLog
	for rows.Next() {
		var log OrphanedLog
		if err := rows.Scan(&log.ID, &log.ContactID, &log.Content, &log.Channel, &log.Timestamp, &log.RetryCount); err != nil {
			slog.Error("OutboxSweeper: Failed to scan log", "error", err)
			continue
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// processOrphanedLog attempts to re-send the notification or mark as FAILED.
func (s *OutboxSweeper) processOrphanedLog(ctx context.Context, log OrphanedLog) error {
	// Check retry limit
	if log.RetryCount >= s.config.MaxRetries {
		slog.Warn("OutboxSweeper: Max retries exceeded, marking FAILED",
			"log_id", log.ID,
			"retries", log.RetryCount,
		)
		return s.updateStatus(ctx, log.ID, SendStatusFailed, log.RetryCount)
	}

	// Fetch contact to determine notification channel
	contact, err := s.getContactByID(ctx, log.ContactID)
	if err != nil {
		slog.Error("OutboxSweeper: Contact lookup failed, marking FAILED",
			"log_id", log.ID,
			"contact_id", log.ContactID,
			"error", err,
		)
		return s.updateStatus(ctx, log.ID, SendStatusFailed, log.RetryCount)
	}

	// Attempt to re-send notification
	sendErr := s.sendNotification(contact, log.Content)
	if sendErr != nil {
		slog.Warn("OutboxSweeper: Send failed, incrementing retry count",
			"log_id", log.ID,
			"retry", log.RetryCount+1,
			"error", sendErr,
		)
		// Increment retry count but keep as PENDING for next sweep
		return s.incrementRetryCount(ctx, log.ID, log.RetryCount+1)
	}

	// Success - mark as SENT
	slog.Info("OutboxSweeper: Successfully recovered orphaned log",
		"log_id", log.ID,
		"contact_id", log.ContactID,
	)
	return s.updateStatus(ctx, log.ID, SendStatusSent, log.RetryCount)
}

// getContactByID fetches a contact by its ID.
func (s *OutboxSweeper) getContactByID(ctx context.Context, contactID uuid.UUID) (*types.Contact, error) {
	query := `
		SELECT id, name, company, COALESCE(phone, ''), COALESCE(email, ''), role, contact_preference
		FROM contacts
		WHERE id = $1
	`
	var contact types.Contact
	var role, preference string
	err := s.db.QueryRow(ctx, query, contactID).Scan(
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

// sendNotification sends via the appropriate channel based on contact preference.
func (s *OutboxSweeper) sendNotification(contact *types.Contact, message string) error {
	switch contact.ContactPreference {
	case types.ContactPreferenceSMS:
		return s.notifier.SendSMS(contact.ID.String(), message)
	case types.ContactPreferenceEmail:
		return s.notifier.SendEmail(contact.Email, "FutureBuild: Notification", message)
	case types.ContactPreferenceBoth:
		if err := s.notifier.SendSMS(contact.ID.String(), message); err != nil {
			slog.Warn("OutboxSweeper: SMS send failed, trying email", "error", err)
		}
		return s.notifier.SendEmail(contact.Email, "FutureBuild: Notification", message)
	default:
		return s.notifier.SendSMS(contact.ID.String(), message)
	}
}

// updateStatus updates the send_status of a communication log.
func (s *OutboxSweeper) updateStatus(ctx context.Context, logID uuid.UUID, status SendStatus, retryCount int) error {
	query := `
		UPDATE communication_logs 
		SET send_status = $1, 
		    metadata = COALESCE(metadata, '{}'::jsonb) || jsonb_build_object('retry_count', $2, 'last_retry', $3)
		WHERE id = $4
	`
	_, err := s.db.Exec(ctx, query, string(status), retryCount, s.clock.Now(), logID)
	return err
}

// incrementRetryCount updates the retry count without changing status.
func (s *OutboxSweeper) incrementRetryCount(ctx context.Context, logID uuid.UUID, newCount int) error {
	query := `
		UPDATE communication_logs 
		SET metadata = COALESCE(metadata, '{}'::jsonb) || jsonb_build_object('retry_count', $1, 'last_retry', $2)
		WHERE id = $3
	`
	_, err := s.db.Exec(ctx, query, newCount, s.clock.Now(), logID)
	return err
}
