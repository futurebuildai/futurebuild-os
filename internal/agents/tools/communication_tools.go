package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/google/uuid"
)

// RegisterCommunicationTools registers tools for contacts, notifications, and messaging.
func RegisterCommunicationTools(r *Registry, directory service.DirectoryServicer, notifier service.NotificationServicer) {
	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "get_contact_for_phase",
			Description: "Look up the subcontractor or contact assigned to a specific construction phase (e.g., '7' for Framing, '9' for Rough-Ins). Returns contact name, company, phone, email, and preference.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"phase_code":{"type":"string","description":"WBS major phase code (e.g., '7' for Framing, '9' for Rough-Ins, '10' for Insulation & Drywall)"}},"required":["phase_code"]}`),
		},
		Handler: func(ctx context.Context, input json.RawMessage) (string, error) {
			scope := MustGetScope(ctx)
			var params struct {
				PhaseCode string `json:"phase_code"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return "", fmt.Errorf("parse input: %w", err)
			}
			contact, err := directory.GetContactForPhase(ctx, scope.ProjectID, scope.OrgID, params.PhaseCode)
			if err != nil {
				return fmt.Sprintf(`{"error":"no contact found for phase %s"}`, params.PhaseCode), nil
			}
			b, _ := json.Marshal(contact)
			return string(b), nil
		},
	})

	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "get_project_manager",
			Description: "Look up the project manager contact for the current project.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{},"required":[]}`),
		},
		Handler: func(ctx context.Context, _ json.RawMessage) (string, error) {
			scope := MustGetScope(ctx)
			contact, err := directory.GetProjectManager(ctx, scope.ProjectID, scope.OrgID)
			if err != nil {
				return `{"error":"no project manager found"}`, nil
			}
			b, _ := json.Marshal(contact)
			return string(b), nil
		},
	})

	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "send_sms",
			Description: "Send an SMS message to a contact. Use draft_message first to compose the message, then send after user approval.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"contact_id":{"type":"string","description":"UUID of the contact to message"},"message":{"type":"string","description":"SMS message body (keep under 160 chars for single SMS)"}},"required":["contact_id","message"]}`),
		},
		Handler: func(ctx context.Context, input json.RawMessage) (string, error) {
			var params struct {
				ContactID string `json:"contact_id"`
				Message   string `json:"message"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return "", fmt.Errorf("parse input: %w", err)
			}
			if err := notifier.SendSMS(params.ContactID, params.Message); err != nil {
				return "", fmt.Errorf("send sms: %w", err)
			}
			return fmt.Sprintf(`{"success":true,"contact_id":"%s","message_length":%d}`, params.ContactID, len(params.Message)), nil
		},
	})

	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "send_email",
			Description: "Send an email to a recipient. Use draft_message first to compose the email, then send after user approval.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"to":{"type":"string","description":"Recipient email address"},"subject":{"type":"string","description":"Email subject line"},"body":{"type":"string","description":"Email body (supports markdown formatting)"}},"required":["to","subject","body"]}`),
		},
		Handler: func(ctx context.Context, input json.RawMessage) (string, error) {
			var params struct {
				To      string `json:"to"`
				Subject string `json:"subject"`
				Body    string `json:"body"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return "", fmt.Errorf("parse input: %w", err)
			}
			if err := notifier.SendEmail(params.To, params.Subject, params.Body); err != nil {
				return "", fmt.Errorf("send email: %w", err)
			}
			return fmt.Sprintf(`{"success":true,"to":"%s","subject":"%s"}`, params.To, params.Subject), nil
		},
	})

	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "draft_message",
			Description: "Draft a curated message (email or SMS) for the user to review before sending. Use this when the user needs help composing a professional response — for example, responding to an upset client, explaining a delay, or negotiating a change order. Present the draft to the user and wait for their approval before using send_email or send_sms.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"channel":{"type":"string","enum":["email","sms"],"description":"Communication channel"},"to_name":{"type":"string","description":"Recipient's name for context"},"to_address":{"type":"string","description":"Email address or phone number"},"subject":{"type":"string","description":"Email subject (required for email, omit for SMS)"},"body":{"type":"string","description":"The drafted message body"},"context":{"type":"string","description":"Brief explanation of why this message is being sent"}},"required":["channel","to_name","to_address","body","context"]}`),
		},
		Handler: func(ctx context.Context, input json.RawMessage) (string, error) {
			// draft_message is a "display" tool — it doesn't execute side effects.
			// It returns the draft back so the chat UI can render it for user review.
			var params struct {
				Channel   string `json:"channel"`
				ToName    string `json:"to_name"`
				ToAddress string `json:"to_address"`
				Subject   string `json:"subject"`
				Body      string `json:"body"`
				Context   string `json:"context"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return "", fmt.Errorf("parse input: %w", err)
			}
			// Return the draft as structured data for the chat UI to render
			result := map[string]interface{}{
				"draft":      true,
				"channel":    params.Channel,
				"to_name":    params.ToName,
				"to_address": params.ToAddress,
				"subject":    params.Subject,
				"body":       params.Body,
				"context":    params.Context,
			}
			b, _ := json.Marshal(result)
			return string(b), nil
		},
	})

	// get_communication_history is registered separately as it needs DB access
}

// RegisterCommunicationHistoryTool registers the communication history lookup tool.
// Separated because it needs direct DB access rather than a service interface.
func RegisterCommunicationHistoryTool(r *Registry, queryFn func(ctx context.Context, contactID uuid.UUID, limit int) (string, error)) {
	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "get_communication_history",
			Description: "Fetch recent communication history with a contact. Returns the last N messages (SMS and email) exchanged with this contact, useful for understanding context before drafting a response.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"contact_id":{"type":"string","description":"UUID of the contact"},"limit":{"type":"integer","description":"Max number of messages to return (default: 10)","default":10}},"required":["contact_id"]}`),
		},
		Handler: func(ctx context.Context, input json.RawMessage) (string, error) {
			var params struct {
				ContactID string `json:"contact_id"`
				Limit     int    `json:"limit"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return "", fmt.Errorf("parse input: %w", err)
			}
			if params.Limit <= 0 {
				params.Limit = 10
			}
			contactID, err := uuid.Parse(params.ContactID)
			if err != nil {
				return "", fmt.Errorf("invalid contact_id: %w", err)
			}
			return queryFn(ctx, contactID, params.Limit)
		},
	})
}
