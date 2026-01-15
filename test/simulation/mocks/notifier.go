// Package mocks provides test doubles for time-travel simulation.
// See PRODUCTION_PLAN.md Step 49 Amendment 3
package mocks

import "sync"

// SpyNotifier captures messages instead of sending them.
// This enables verification of notification timing without side effects.
type SpyNotifier struct {
	mu         sync.Mutex
	SentSMS    []string // Format: "To:Body"
	SentEmails []string // Format: "To:Subject:Body"
}

// SendSMS captures the SMS message in SentSMS slice.
func (s *SpyNotifier) SendSMS(contactID string, message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SentSMS = append(s.SentSMS, contactID+":"+message)
	return nil
}

// SendEmail captures the email in SentEmails slice.
func (s *SpyNotifier) SendEmail(to string, subject string, body string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SentEmails = append(s.SentEmails, to+":"+subject+":"+body)
	return nil
}

// GetSMSCount returns the number of SMS messages sent.
func (s *SpyNotifier) GetSMSCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.SentSMS)
}

// GetEmailCount returns the number of emails sent.
func (s *SpyNotifier) GetEmailCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.SentEmails)
}

// ContainsSMS checks if any sent SMS contains the given substring.
func (s *SpyNotifier) ContainsSMS(substring string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, sms := range s.SentSMS {
		if Contains(sms, substring) {
			return true
		}
	}
	return false
}

// Contains is a helper function for substring matching.
func Contains(s, substr string) bool {
	return len(substr) <= len(s) && containsHelper(s, substr)
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
