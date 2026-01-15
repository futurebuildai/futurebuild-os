package handlers

import (
	"log/slog"
	"net/http"

	"github.com/colton/futurebuild/internal/agents"
	"github.com/colton/futurebuild/internal/api/response"
)

// WebhookHandler handles inbound webhooks from external services (Twilio, SendGrid, etc.)
// See PRODUCTION_PLAN.md Step 47 (Sub Liaison Agent API Integration)
type WebhookHandler struct {
	liaisonAgent *agents.SubLiaisonAgent
}

// NewWebhookHandler creates a new WebhookHandler with the Sub Liaison Agent.
func NewWebhookHandler(liaisonAgent *agents.SubLiaisonAgent) *WebhookHandler {
	return &WebhookHandler{
		liaisonAgent: liaisonAgent,
	}
}

// HandleInboundMessage processes inbound SMS/Email messages from subcontractors.
// POST /webhooks/messages
//
// Expected form values (Twilio format):
//   - From: sender phone or email
//   - Body: message content
//
// TODO: Implement Twilio/SendGrid signature verification
// See PRODUCTION_PLAN.md Step 47 (Security)
func (h *WebhookHandler) HandleInboundMessage(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement Twilio/SendGrid signature verification
	// signature := r.Header.Get("X-Twilio-Signature")
	// if !verifyTwilioSignature(signature, r) {
	//     response.JSONError(w, http.StatusUnauthorized, "Invalid signature")
	//     return
	// }

	if err := r.ParseForm(); err != nil {
		slog.Warn("webhook: failed to parse form", "error", err)
		response.JSONError(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	sender := r.FormValue("From")
	body := r.FormValue("Body")

	if sender == "" || body == "" {
		slog.Warn("webhook: missing required fields", "sender", sender, "body_len", len(body))
		response.JSONError(w, http.StatusBadRequest, "Missing From or Body field")
		return
	}

	slog.Info("webhook: inbound message received",
		"sender", sender,
		"body_length", len(body),
	)

	// Delegate to SubLiaisonAgent (async-safe, errors logged internally)
	if err := h.liaisonAgent.HandleInboundMessage(r.Context(), sender, body); err != nil {
		slog.Error("webhook: liaison agent error", "error", err)
		// Still return 200 to ACK the webhook (Twilio retries on non-2xx)
	}

	// Twilio expects TwiML response or 200 OK
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
