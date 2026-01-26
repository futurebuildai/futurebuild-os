package service

import (
	"log/slog"

	"github.com/colton/futurebuild/pkg/types"
)

// CompositeNotificationProvider combines email and SMS providers into a single NotificationService.
// Uses SendGrid for email and Twilio for SMS. See LAUNCH_PLAN.md P2.
type CompositeNotificationProvider struct {
	emailProvider types.NotificationService
	smsProvider   types.NotificationService
}

// NewCompositeNotificationProvider creates a notification provider that routes
// email to SendGrid and SMS to Twilio.
//
// If either provider is nil, it falls back to ConsoleEmailProvider for that channel.
func NewCompositeNotificationProvider(
	emailProvider types.NotificationService,
	smsProvider types.NotificationService,
) *CompositeNotificationProvider {
	fallback := NewConsoleEmailProvider()

	if emailProvider == nil {
		emailProvider = fallback
	}
	if smsProvider == nil {
		smsProvider = fallback
	}

	return &CompositeNotificationProvider{
		emailProvider: emailProvider,
		smsProvider:   smsProvider,
	}
}

// SendEmail routes to the email provider (SendGrid or Console).
func (c *CompositeNotificationProvider) SendEmail(to string, subject string, body string) error {
	return c.emailProvider.SendEmail(to, subject, body)
}

// SendSMS routes to the SMS provider (Twilio or Console).
func (c *CompositeNotificationProvider) SendSMS(contactID string, message string) error {
	return c.smsProvider.SendSMS(contactID, message)
}

// NewNotificationService creates the appropriate notification service based on configuration.
// This is a factory function that simplifies server.go wiring.
//
// Configuration logic:
// - If SendGrid API key is set, use SendGrid for email
// - If Twilio credentials are set, use Twilio for SMS
// - Otherwise, fall back to console logging (development mode)
func NewNotificationService(
	sendGridAPIKey, emailFromAddress, emailFromName string,
	twilioAccountSID, twilioAuthToken, twilioFromNumber string,
) types.NotificationService {
	var emailProvider types.NotificationService
	var smsProvider types.NotificationService

	// Configure email provider
	if sendGridAPIKey != "" {
		emailProvider = NewSendGridProvider(sendGridAPIKey, emailFromAddress, emailFromName)
		slog.Info("notification: using SendGrid for email")
	} else {
		emailProvider = NewConsoleEmailProvider()
		slog.Info("notification: using Console for email (development mode)")
	}

	// Configure SMS provider
	if twilioAccountSID != "" && twilioAuthToken != "" && twilioFromNumber != "" {
		smsProvider = NewTwilioProvider(twilioAccountSID, twilioAuthToken, twilioFromNumber)
		slog.Info("notification: using Twilio for SMS")
	} else {
		smsProvider = NewConsoleEmailProvider()
		slog.Info("notification: using Console for SMS (development mode)")
	}

	return NewCompositeNotificationProvider(emailProvider, smsProvider)
}
