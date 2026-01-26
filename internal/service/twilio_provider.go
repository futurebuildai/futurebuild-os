package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

// TwilioProvider implements types.NotificationService using Twilio API.
// See LAUNCH_PLAN.md P2 (Notifications/Toast UI + Field Portal).
type TwilioProvider struct {
	accountSID string
	authToken  string
	fromNumber string
	httpClient *http.Client
}

// NewTwilioProvider creates a new Twilio SMS provider.
// Returns nil if accountSID is empty (use ConsoleEmailProvider instead for development).
func NewTwilioProvider(accountSID, authToken, fromNumber string) *TwilioProvider {
	if accountSID == "" {
		return nil
	}
	return &TwilioProvider{
		accountSID: accountSID,
		authToken:  authToken,
		fromNumber: fromNumber,
		httpClient: &http.Client{},
	}
}

// twilioResponse represents the Twilio API response for message creation.
type twilioResponse struct {
	SID          string `json:"sid"`
	Status       string `json:"status"`
	ErrorCode    *int   `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// SendSMS sends an SMS message via Twilio API.
// contactID should be a phone number in E.164 format (e.g., +1234567890).
func (p *TwilioProvider) SendSMS(contactID string, message string) error {
	apiURL := fmt.Sprintf(
		"https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json",
		p.accountSID,
	)

	// Build form data
	data := url.Values{}
	data.Set("To", contactID)
	data.Set("From", p.fromNumber)
	data.Set("Body", message)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		slog.Error("twilio: failed to create request", "error", err)
		return fmt.Errorf("failed to create SMS request: %w", err)
	}

	req.SetBasicAuth(p.accountSID, p.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		slog.Error("twilio: failed to send request", "error", err, "to", contactID)
		return fmt.Errorf("failed to send SMS: %w", err)
	}
	defer resp.Body.Close()

	var twilioResp twilioResponse
	if err := json.NewDecoder(resp.Body).Decode(&twilioResp); err != nil {
		slog.Error("twilio: failed to decode response", "error", err)
		return fmt.Errorf("failed to decode Twilio response: %w", err)
	}

	if resp.StatusCode >= 400 {
		slog.Error("twilio: API error",
			"status", resp.StatusCode,
			"error_code", twilioResp.ErrorCode,
			"error_message", twilioResp.ErrorMessage,
			"to", contactID,
		)
		return fmt.Errorf("twilio API error: %s (code: %v)", twilioResp.ErrorMessage, twilioResp.ErrorCode)
	}

	slog.Info("twilio: SMS sent successfully",
		"to", contactID,
		"sid", twilioResp.SID,
		"status", twilioResp.Status,
	)
	return nil
}

// SendEmail is not implemented via Twilio. Use SendGrid for email.
// This method delegates to the fallback behavior (no-op with warning).
func (p *TwilioProvider) SendEmail(to string, subject string, body string) error {
	slog.Warn("twilio: email not implemented, use SendGrid instead", "to", to, "subject", subject)
	return nil
}
