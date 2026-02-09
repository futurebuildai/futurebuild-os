package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// BirdProvider implements types.NotificationService using Bird (MessageBird) API.
// Unified provider for both SMS and Email, replacing Twilio + Resend.
// See https://developers.messagebird.com/api/
type BirdProvider struct {
	accessKey   string
	originator  string // SMS sender ID (alphanumeric or phone number)
	fromAddress string // Email sender address
	fromName    string // Email sender name
	httpClient  *http.Client
}

// NewBirdProvider creates a new Bird notification provider.
// Returns nil if accessKey is empty (use ConsoleEmailProvider instead for development).
func NewBirdProvider(accessKey, originator, fromAddress, fromName string) *BirdProvider {
	if accessKey == "" {
		return nil
	}
	return &BirdProvider{
		accessKey:   accessKey,
		originator:  originator,
		fromAddress: fromAddress,
		fromName:    fromName,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// birdSMSResponse represents the MessageBird SMS API response.
type birdSMSResponse struct {
	ID     string `json:"id"`
	Status string `json:"status,omitempty"`
	Errors []struct {
		Code        int    `json:"code"`
		Description string `json:"description"`
		Parameter   string `json:"parameter,omitempty"`
	} `json:"errors,omitempty"`
}

// SendSMS sends an SMS message via Bird (MessageBird) API.
// contactID should be a phone number in E.164 format (e.g., +1234567890).
func (p *BirdProvider) SendSMS(contactID string, message string) error {
	apiURL := "https://rest.messagebird.com/messages"

	// Build form data (MessageBird SMS API uses form-urlencoded)
	data := url.Values{}
	data.Set("recipients", contactID)
	data.Set("originator", p.originator)
	data.Set("body", message)

	req, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		slog.Error("bird: failed to create SMS request", "error", err)
		return fmt.Errorf("failed to create SMS request: %w", err)
	}

	req.Header.Set("Authorization", "AccessKey "+p.accessKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		slog.Error("bird: failed to send SMS request", "error", err, "to", contactID)
		return fmt.Errorf("failed to send SMS: %w", err)
	}
	defer resp.Body.Close()

	var birdResp birdSMSResponse
	if err := json.NewDecoder(resp.Body).Decode(&birdResp); err != nil {
		slog.Error("bird: failed to decode SMS response", "error", err)
		return fmt.Errorf("failed to decode Bird SMS response: %w", err)
	}

	if resp.StatusCode >= 400 {
		errMsg := "unknown error"
		if len(birdResp.Errors) > 0 {
			errMsg = birdResp.Errors[0].Description
		}
		slog.Error("bird: SMS API error",
			"status", resp.StatusCode,
			"error", errMsg,
			"to", contactID,
		)
		return fmt.Errorf("bird SMS API error: %s (status: %d)", errMsg, resp.StatusCode)
	}

	slog.Info("bird: SMS sent successfully",
		"to", contactID,
		"id", birdResp.ID,
	)
	return nil
}

// birdEmailRequest represents the Bird Email API request payload.
type birdEmailRequest struct {
	From struct {
		Email string `json:"email"`
		Name  string `json:"name,omitempty"`
	} `json:"from"`
	To []struct {
		Email string `json:"email"`
	} `json:"to"`
	Subject  string `json:"subject"`
	TextBody string `json:"text,omitempty"`
	HTMLBody string `json:"html,omitempty"`
}

// birdEmailResponse represents the Bird Email API response.
type birdEmailResponse struct {
	ID     string `json:"id"`
	Status string `json:"status,omitempty"`
	Errors []struct {
		Code        int    `json:"code"`
		Description string `json:"description"`
	} `json:"errors,omitempty"`
}

// SendEmail sends an email via Bird Email API.
func (p *BirdProvider) SendEmail(to string, subject string, body string) error {
	apiURL := "https://email.messagebird.com/v1/send"

	// Build HTML content with basic formatting (matching Resend provider)
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

	reqBody := birdEmailRequest{
		Subject:  subject,
		TextBody: body,
		HTMLBody: htmlContent,
	}
	reqBody.From.Email = p.fromAddress
	reqBody.From.Name = p.fromName
	reqBody.To = []struct {
		Email string `json:"email"`
	}{{Email: to}}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		slog.Error("bird: failed to marshal email request", "error", err)
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		slog.Error("bird: failed to create email request", "error", err)
		return fmt.Errorf("failed to create email request: %w", err)
	}

	req.Header.Set("Authorization", "AccessKey "+p.accessKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		slog.Error("bird: failed to send email request", "error", err, "to", to)
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	var birdResp birdEmailResponse
	if err := json.NewDecoder(resp.Body).Decode(&birdResp); err != nil {
		slog.Error("bird: failed to decode email response", "error", err)
		return fmt.Errorf("failed to decode Bird email response: %w", err)
	}

	if resp.StatusCode >= 400 {
		errMsg := "unknown error"
		if len(birdResp.Errors) > 0 {
			errMsg = birdResp.Errors[0].Description
		}
		slog.Error("bird: email API error",
			"status", resp.StatusCode,
			"error", errMsg,
			"to", to,
		)
		return fmt.Errorf("bird email API error: %s (status: %d)", errMsg, resp.StatusCode)
	}

	slog.Info("bird: email sent successfully",
		"to", to,
		"subject", subject,
		"id", birdResp.ID,
	)
	return nil
}
