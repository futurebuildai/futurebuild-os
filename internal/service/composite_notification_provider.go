package service

import (
	"log/slog"

	"github.com/colton/futurebuild/pkg/types"
)

// CompositeNotificationProvider combines email and SMS providers into a single NotificationService.
// Uses Resend for email and Twilio for SMS. See LAUNCH_PLAN.md P2.
type CompositeNotificationProvider struct {
	emailProvider types.NotificationService
	smsProvider   types.NotificationService
}

// NewCompositeNotificationProvider creates a notification provider that routes
// email to Resend and SMS to Twilio.
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

// SendEmail routes to the email provider (Resend or Console).
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
// Provider logic (controlled by NOTIFICATION_PROVIDER env var):
// - "bird": Use Bird (MessageBird) for both SMS and Email
// - "legacy": Use Resend for email, Twilio for SMS (original behavior)
// - "console" or default: Fall back to console logging (development mode)
func NewNotificationService(
	resendAPIKey string,
	emailFromAddress, emailFromName string,
	twilioAccountSID, twilioAuthToken, twilioFromNumber string,
	birdAccessKey, birdOriginator string,
	provider string,
) types.NotificationService {
	switch provider {
	case "bird":
		if birdAccessKey == "" {
			slog.Warn("notification: NOTIFICATION_PROVIDER=bird but BIRD_ACCESS_KEY not set, falling back to console")
			return NewConsoleEmailProvider()
		}
		slog.Info("notification: using Bird for SMS and email")
		return NewBirdProvider(birdAccessKey, birdOriginator, emailFromAddress, emailFromName)

	case "legacy":
		return newLegacyNotificationService(
			resendAPIKey, emailFromAddress, emailFromName,
			twilioAccountSID, twilioAuthToken, twilioFromNumber,
		)

	default:
		slog.Info("notification: using Console (development mode)")
		return NewConsoleEmailProvider()
	}
}

// newLegacyNotificationService creates the composite Resend + Twilio provider.
func newLegacyNotificationService(
	resendAPIKey string,
	emailFromAddress, emailFromName string,
	twilioAccountSID, twilioAuthToken, twilioFromNumber string,
) types.NotificationService {
	var emailProvider types.NotificationService
	var smsProvider types.NotificationService

	// Configure email provider: Resend > Console
	if resendAPIKey != "" {
		emailProvider = NewResendProvider(resendAPIKey, emailFromAddress, emailFromName)
		slog.Info("notification: using Resend for email")
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
