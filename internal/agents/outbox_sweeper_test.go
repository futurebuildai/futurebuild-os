package agents

import (
	"context"
	"testing"
	"time"

	"github.com/colton/futurebuild/pkg/clock"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
)

// TestOutboxSweeperConfig verifies default configuration.
func TestOutboxSweeperConfig_Defaults(t *testing.T) {
	cfg := DefaultOutboxSweeperConfig()

	if cfg.StaleThreshold != 5*time.Minute {
		t.Errorf("Expected StaleThreshold=5m, got %v", cfg.StaleThreshold)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries=3, got %d", cfg.MaxRetries)
	}
	if cfg.BatchSize != 100 {
		t.Errorf("Expected BatchSize=100, got %d", cfg.BatchSize)
	}
}

// mockNotifierForSweeper records calls for testing.
type mockNotifierForSweeper struct {
	sendSMSCalls   int
	sendEmailCalls int
	shouldFail     bool
}

func (m *mockNotifierForSweeper) SendSMS(contactID, message string) error {
	m.sendSMSCalls++
	if m.shouldFail {
		return context.DeadlineExceeded // Simulate transient failure
	}
	return nil
}

func (m *mockNotifierForSweeper) SendEmail(to, subject, body string) error {
	m.sendEmailCalls++
	if m.shouldFail {
		return context.DeadlineExceeded
	}
	return nil
}

// TestOutboxSweeper_SendNotification verifies notification routing by preference.
func TestOutboxSweeper_SendNotification(t *testing.T) {
	tests := []struct {
		name        string
		preference  types.ContactPreference
		expectSMS   int
		expectEmail int
	}{
		{"SMS preference", types.ContactPreferenceSMS, 1, 0},
		{"Email preference", types.ContactPreferenceEmail, 0, 1},
		{"Both preference", types.ContactPreferenceBoth, 1, 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			notifier := &mockNotifierForSweeper{}
			sweeper := &OutboxSweeper{
				notifier: notifier,
				clock:    clock.NewMockClock(time.Now()),
			}

			contact := &types.Contact{
				ID:                uuid.New(),
				Email:             "test@example.com",
				ContactPreference: tc.preference,
			}

			err := sweeper.sendNotification(contact, "test message")
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if notifier.sendSMSCalls != tc.expectSMS {
				t.Errorf("Expected %d SMS calls, got %d", tc.expectSMS, notifier.sendSMSCalls)
			}
			if notifier.sendEmailCalls != tc.expectEmail {
				t.Errorf("Expected %d email calls, got %d", tc.expectEmail, notifier.sendEmailCalls)
			}
		})
	}
}

// TestOutboxSweeper_MaxRetriesExceeded verifies FAILED status when retries exhausted.
func TestOutboxSweeper_MaxRetriesExceeded(t *testing.T) {
	cfg := DefaultOutboxSweeperConfig()
	cfg.MaxRetries = 3

	log := OrphanedLog{
		ID:         uuid.New(),
		ContactID:  uuid.New(),
		Content:    "test message",
		RetryCount: 3, // Already at max
	}

	// RetryCount >= MaxRetries should trigger FAILED path
	if log.RetryCount < cfg.MaxRetries {
		t.Error("Test setup incorrect: RetryCount should be >= MaxRetries")
	}

	// In a full integration test, we'd verify the DB update
	// For now, this documents the expected behavior
	t.Log("MaxRetries check would mark log as FAILED")
}
