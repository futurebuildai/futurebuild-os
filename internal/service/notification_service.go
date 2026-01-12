package service

import (
	"log"
)

// ConsoleEmailProvider implements types.NotificationService by logging to stdout.
// See Technical Specification: Magic Link Authentication (Stateful) Section 2.1
type ConsoleEmailProvider struct{}

func NewConsoleEmailProvider() *ConsoleEmailProvider {
	return &ConsoleEmailProvider{}
}

func (p *ConsoleEmailProvider) SendEmail(to string, subject string, body string) error {
	// REQUIRED: Log to stdout for Step 21 verification
	log.Printf("[MOCK EMAIL] To: %s | Subject: %s | Body: %s", to, subject, body)
	return nil
}

func (p *ConsoleEmailProvider) SendSMS(contactID string, message string) error {
	log.Printf("[MOCK SMS] To ContactID: %s | Message: %s", contactID, message)
	return nil
}
