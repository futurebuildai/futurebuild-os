package agents

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
)

// ErrNotFound is a sentinel error for testing.
var errNotFound = errors.New("not found")

// mockDirectoryService implements DirectoryService for testing.
type mockDirectoryService struct {
	contact *types.Contact
	err     error
}

func (m *mockDirectoryService) GetContactForPhase(_ context.Context, _, _ uuid.UUID, _ string) (*types.Contact, error) {
	return m.contact, m.err
}

// mockNotificationService implements NotificationService for testing.
type mockNotificationService struct {
	smsSent    []string
	emailsSent []string
	err        error
}

func (m *mockNotificationService) SendSMS(contactID, message string) error {
	m.smsSent = append(m.smsSent, contactID+":"+message)
	return m.err
}

func (m *mockNotificationService) SendEmail(to, subject, body string) error {
	m.emailsSent = append(m.emailsSent, to+":"+subject)
	return m.err
}

// TestConfirmArrival_NoContactFound is an integration test that requires a mock DB.
// For unit tests, we verify the helper functions below.
// Full integration tests are in test/integration/sub_liaison_test.go
// See PRODUCTION_PLAN.md Step 47 (Testing Strategy)

// TestParsePercentage verifies percentage extraction from various message formats.
func TestParsePercentage(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		found    bool
	}{
		{"50%", 50, true},
		{"we are 75 % done", 75, true},
		{"done", 100, true},
		{"complete", 100, true},
		{"finished", 100, true},
		{"25 percent complete", 25, true},
		{"hello world", 0, false},
		{"no update", 0, false},
		{"150%", 0, false}, // Invalid percentage
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, found := parsePercentage(tc.input)
			if found != tc.found {
				t.Errorf("parsePercentage(%q) found=%v, want found=%v", tc.input, found, tc.found)
			}
			if found && result != tc.expected {
				t.Errorf("parsePercentage(%q) = %d, want %d", tc.input, result, tc.expected)
			}
		})
	}
}

// TestContainsDelayIndicator verifies delay keyword detection.
func TestContainsDelayIndicator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"we will be delayed", true},
		{"running late", true},
		{"no, can't make it", true},
		{"waiting for materials", true},
		{"on my way", false},
		{"yes confirmed", false},
		{"50% complete", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := containsDelayIndicator(tc.input)
			if result != tc.expected {
				t.Errorf("containsDelayIndicator(%q) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}

// TestExtractImageURL verifies image URL extraction from messages.
func TestExtractImageURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"check this https://example.com/photo.jpg", "https://example.com/photo.jpg"},
		{"image: http://cdn.site.com/img.png uploaded", "http://cdn.site.com/img.png"},
		{"no image here", ""},
		{"just text https://example.com/page", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := extractImageURL(tc.input)
			if result != tc.expected {
				t.Errorf("extractImageURL(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestExtractPhaseCode verifies WBS code to phase code mapping.
func TestExtractPhaseCode(t *testing.T) {
	tests := []struct {
		wbsCode  string
		expected string
	}{
		{"9.1.2", "9"},
		{"14.3", "14"},
		{"5", "5"},
		{"", ""},
	}

	for _, tc := range tests {
		t.Run(tc.wbsCode, func(t *testing.T) {
			result := extractPhaseCode(tc.wbsCode)
			if result != tc.expected {
				t.Errorf("extractPhaseCode(%q) = %q, want %q", tc.wbsCode, result, tc.expected)
			}
		})
	}
}

// TestFormatDate verifies date formatting with nil handling.
func TestFormatDate(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    *time.Time
		expected string
	}{
		{"nil date", nil, "TBD"},
		{"valid date", &now, "Jan 15, 2026"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatDate(tc.input)
			if result != tc.expected {
				t.Errorf("formatDate = %q, want %q", result, tc.expected)
			}
		})
	}
}

// TestTruncateString verifies string truncation.
func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a very long string", 15, "this is a ve..."},
		{"", 5, ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := truncateString(tc.input, tc.maxLen)
			if result != tc.expected {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tc.input, tc.maxLen, result, tc.expected)
			}
		})
	}
}
