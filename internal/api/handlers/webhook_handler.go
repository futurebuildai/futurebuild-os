package handlers

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"

	"github.com/colton/futurebuild/internal/agents"
	"github.com/colton/futurebuild/internal/api/response"
)

// WebhookHandler handles inbound webhooks from external services (Twilio, SendGrid).
// See PRODUCTION_PLAN.md Step 48 (Inbound Message Processing)
type WebhookHandler struct {
	processor     *agents.InboundProcessor
	webhookSecret string
}

// NewWebhookHandler creates a new WebhookHandler with the Inbound Processor.
// See PRODUCTION_PLAN.md Step 48
func NewWebhookHandler(processor *agents.InboundProcessor, webhookSecret string) *WebhookHandler {
	return &WebhookHandler{
		processor:     processor,
		webhookSecret: webhookSecret,
	}
}

// HandleSMS processes inbound SMS messages from Twilio.
// POST /api/v1/webhooks/sms
//
// Expected form values (Twilio format):
//   - From: sender phone number
//   - Body: message content
//   - MessageSid: unique message identifier (idempotency key)
//
// See PRODUCTION_PLAN.md Step 48
func (h *WebhookHandler) HandleSMS(w http.ResponseWriter, r *http.Request) {
	// Step 1: Signature Verification
	// See L7 Security Amendment
	if h.webhookSecret != "" {
		// Prevent DoS: Limit body reading to 1MB
		r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Warn("webhook/sms: failed to read body", "error", err)
			response.JSONError(w, http.StatusBadRequest, "Failed to read request body")
			return
		}

		signature := r.Header.Get("X-FutureBuild-Signature")
		if !agents.VerifySignature(bodyBytes, signature, h.webhookSecret) {
			slog.Warn("webhook/sms: invalid signature",
				"remote_addr", r.RemoteAddr,
			)
			response.JSONError(w, http.StatusUnauthorized, "Invalid signature")
			return
		}

		// Restore body for ParseForm
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	// Step 2: Parse Form Data
	if err := r.ParseForm(); err != nil {
		slog.Warn("webhook/sms: failed to parse form", "error", err)
		response.JSONError(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	sender := r.FormValue("From")
	body := r.FormValue("Body")
	messageSid := r.FormValue("MessageSid")

	if sender == "" || body == "" {
		slog.Warn("webhook/sms: missing required fields",
			"sender", sender,
			"body_len", len(body),
		)
		response.JSONError(w, http.StatusBadRequest, "Missing From or Body field")
		return
	}

	slog.Info("webhook/sms: inbound message received",
		"sender", sender,
		"body_length", len(body),
		"message_sid", messageSid,
	)

	// Step 3: Build normalized message
	msg := agents.InboundMessage{
		ExternalID: messageSid,
		Sender:     sender,
		Body:       body,
		ImageURLs:  agents.ExtractImageURLs(body),
		Channel:    "SMS",
	}

	// Step 4: Process via InboundProcessor
	if err := h.processor.ProcessIncoming(r.Context(), msg); err != nil {
		slog.Error("webhook/sms: processor error", "error", err)
		// Still return 200 to ACK the webhook (Twilio retries on non-2xx)
	}

	// Twilio expects TwiML response or 200 OK
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

// HandleEmail processes inbound email messages from SendGrid.
// POST /api/v1/webhooks/email
//
// Expected form values (SendGrid Inbound Parse format):
//   - from: sender email address
//   - text: message body
//   - headers: contains Message-ID for idempotency
//
// See PRODUCTION_PLAN.md Step 48
func (h *WebhookHandler) HandleEmail(w http.ResponseWriter, r *http.Request) {
	// Step 1: Signature Verification
	if h.webhookSecret != "" {
		// Prevent DoS: Limit body reading to 10MB (emails can be larger)
		r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Warn("webhook/email: failed to read body", "error", err)
			response.JSONError(w, http.StatusBadRequest, "Failed to read request body")
			return
		}

		signature := r.Header.Get("X-FutureBuild-Signature")
		if !agents.VerifySignature(bodyBytes, signature, h.webhookSecret) {
			slog.Warn("webhook/email: invalid signature",
				"remote_addr", r.RemoteAddr,
			)
			response.JSONError(w, http.StatusUnauthorized, "Invalid signature")
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	// Step 2: Parse Form Data (SendGrid Inbound Parse uses multipart/form-data)
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		// Fallback to regular form
		if err := r.ParseForm(); err != nil {
			slog.Warn("webhook/email: failed to parse form", "error", err)
			response.JSONError(w, http.StatusBadRequest, "Invalid form data")
			return
		}
	}

	sender := r.FormValue("from")
	body := r.FormValue("text")
	messageID := r.FormValue("Message-ID")
	if messageID == "" {
		// Try extracting from headers
		messageID = extractMessageIDFromHeaders(r.FormValue("headers"))
	}

	if sender == "" || body == "" {
		slog.Warn("webhook/email: missing required fields",
			"sender", sender,
			"body_len", len(body),
		)
		response.JSONError(w, http.StatusBadRequest, "Missing from or text field")
		return
	}

	slog.Info("webhook/email: inbound message received",
		"sender", sender,
		"body_length", len(body),
		"message_id", messageID,
	)

	// Step 3: Build normalized message
	msg := agents.InboundMessage{
		ExternalID: messageID,
		Sender:     sender,
		Body:       body,
		ImageURLs:  agents.ExtractImageURLs(body),
		Channel:    "Email",
	}

	// Step 4: Process via InboundProcessor
	if err := h.processor.ProcessIncoming(r.Context(), msg); err != nil {
		slog.Error("webhook/email: processor error", "error", err)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

// HandleInboundMessage is the legacy endpoint for backwards compatibility.
// POST /webhooks/messages
// Deprecated: Use /api/v1/webhooks/sms or /api/v1/webhooks/email instead.
func (h *WebhookHandler) HandleInboundMessage(w http.ResponseWriter, r *http.Request) {
	// Redirect to SMS handler for legacy support
	h.HandleSMS(w, r)
}

// --- Helpers ---

func extractMessageIDFromHeaders(headers string) string {
	// Simple extraction - in production would parse RFC 2822 headers
	for _, line := range splitLines(headers) {
		if len(line) > 12 && line[:11] == "Message-ID:" {
			return trimSpace(line[11:])
		}
	}
	return ""
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
