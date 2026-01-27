package service

import (
	"fmt"
	"log/slog"

	"github.com/resend/resend-go/v2"
)

// ResendProvider implements types.NotificationService using Resend API.
// Resend offers a modern API with generous free tier (3,000 emails/month).
// See https://resend.com/docs
type ResendProvider struct {
	client      *resend.Client
	fromAddress string
	fromName    string
}

// NewResendProvider creates a new Resend email provider.
// Returns nil if apiKey is empty (use ConsoleEmailProvider instead for development).
func NewResendProvider(apiKey, fromAddress, fromName string) *ResendProvider {
	if apiKey == "" {
		return nil
	}
	return &ResendProvider{
		client:      resend.NewClient(apiKey),
		fromAddress: fromAddress,
		fromName:    fromName,
	}
}

// SendEmail sends an email via Resend API.
func (p *ResendProvider) SendEmail(to string, subject string, body string) error {
	from := fmt.Sprintf("%s <%s>", p.fromName, p.fromAddress)

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

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Html:    htmlContent,
		Text:    body,
	}

	sent, err := p.client.Emails.Send(params)
	if err != nil {
		slog.Error("resend: failed to send email", "error", err, "to", to)
		return fmt.Errorf("failed to send email: %w", err)
	}

	slog.Info("resend: email sent successfully", "to", to, "subject", subject, "id", sent.Id)
	return nil
}

// SendSMS is not implemented via Resend. Use a dedicated SMS provider (e.g., Twilio).
// This is a no-op that logs the request.
func (p *ResendProvider) SendSMS(contactID string, message string) error {
	slog.Warn("resend: SMS not implemented, skipping", "contact_id", contactID)
	return nil
}
