package service

import (
	"fmt"
	"log/slog"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendGridProvider implements types.NotificationService using SendGrid API.
// See LAUNCH_STRATEGY.md Task A3.
type SendGridProvider struct {
	apiKey      string
	fromAddress string
	fromName    string
}

// NewSendGridProvider creates a new SendGrid email provider.
// Returns nil if apiKey is empty (use ConsoleEmailProvider instead for development).
func NewSendGridProvider(apiKey, fromAddress, fromName string) *SendGridProvider {
	if apiKey == "" {
		return nil
	}
	return &SendGridProvider{
		apiKey:      apiKey,
		fromAddress: fromAddress,
		fromName:    fromName,
	}
}

// SendEmail sends an email via SendGrid API.
func (p *SendGridProvider) SendEmail(to string, subject string, body string) error {
	from := mail.NewEmail(p.fromName, p.fromAddress)
	toEmail := mail.NewEmail("", to)

	// Create plain text content
	plainContent := body

	// Create HTML content with basic formatting
	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .button { display: inline-block; padding: 12px 24px; background-color: #2563eb; color: white; text-decoration: none; border-radius: 6px; margin: 20px 0; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb; font-size: 12px; color: #6b7280; }
    </style>
</head>
<body>
    <div class="container">
        <p>%s</p>
        <div class="footer">
            <p>FutureBuild - AI-Powered Construction Management</p>
        </div>
    </div>
</body>
</html>`, body)

	message := mail.NewSingleEmail(from, subject, toEmail, plainContent, htmlContent)

	client := sendgrid.NewSendClient(p.apiKey)
	response, err := client.Send(message)
	if err != nil {
		slog.Error("sendgrid: failed to send email", "error", err, "to", to)
		return fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode >= 400 {
		slog.Error("sendgrid: API error", "status", response.StatusCode, "body", response.Body, "to", to)
		return fmt.Errorf("sendgrid API error: status %d", response.StatusCode)
	}

	slog.Info("sendgrid: email sent successfully", "to", to, "subject", subject, "status", response.StatusCode)
	return nil
}

// SendSMS is not implemented via SendGrid. Use a dedicated SMS provider (e.g., Twilio).
// For now, this logs the request and returns nil (no-op).
func (p *SendGridProvider) SendSMS(contactID string, message string) error {
	// TODO: Integrate Twilio or another SMS provider
	slog.Warn("sendgrid: SMS not implemented, skipping", "contact_id", contactID)
	return nil
}
