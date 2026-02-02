package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/colton/futurebuild/internal/api/response"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// maxClerkWebhookBodySize is the maximum allowed body size for Clerk webhooks (512KB).
	maxClerkWebhookBodySize = 512 * 1024

	// clerkTimestampTolerance is the maximum age of a webhook event (5 minutes).
	clerkTimestampTolerance = 5 * time.Minute
)

// ClerkWebhookHandler handles Clerk webhook events for org/user sync.
// See PHASE_12_PRD.md Step 80: Organization Manager.
//
// Pattern follows github_webhook_handler.go: read body, verify signature, route by event type.
// Svix verification uses manual HMAC-SHA256 (no new dependency).
type ClerkWebhookHandler struct {
	db            *pgxpool.Pool
	webhookSecret string
}

// NewClerkWebhookHandler creates a new handler with Svix HMAC verification.
func NewClerkWebhookHandler(db *pgxpool.Pool, webhookSecret string) *ClerkWebhookHandler {
	return &ClerkWebhookHandler{
		db:            db,
		webhookSecret: webhookSecret,
	}
}

// HandleClerkWebhook processes incoming Clerk webhook events.
// POST /api/v1/webhooks/clerk
//
// Security: Svix HMAC-SHA256 verification with fail-closed behavior.
// Events handled:
//   - organization.created / organization.updated -> upsert organizations by external_id
//   - organizationMembership.created -> upsert users with org_id + role
func (h *ClerkWebhookHandler) HandleClerkWebhook(w http.ResponseWriter, r *http.Request) {
	// Step 1: Fail-Closed - Reject if secret not configured
	if h.webhookSecret == "" {
		slog.Error("webhook/clerk: webhook secret not configured (fail-closed)")
		response.JSONError(w, http.StatusForbidden, "Webhook not configured")
		return
	}

	// Step 2: Read body with size limit (DoS prevention)
	r.Body = http.MaxBytesReader(w, r.Body, maxClerkWebhookBodySize)
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Warn("webhook/clerk: failed to read body", "error", err)
		response.JSONError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	// Step 3: Verify Svix signature
	svixID := r.Header.Get("svix-id")
	svixTimestamp := r.Header.Get("svix-timestamp")
	svixSignature := r.Header.Get("svix-signature")

	if !h.verifySvixSignature(bodyBytes, svixID, svixTimestamp, svixSignature) {
		slog.Warn("webhook/clerk: invalid Svix signature",
			"remote_addr", r.RemoteAddr,
			"svix_id", svixID,
		)
		response.JSONError(w, http.StatusForbidden, "Invalid signature")
		return
	}

	// Step 4: Parse event envelope
	var envelope clerkWebhookEnvelope
	if err := json.Unmarshal(bodyBytes, &envelope); err != nil {
		slog.Warn("webhook/clerk: invalid JSON payload", "error", err)
		response.JSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	slog.Info("webhook/clerk: received event",
		"type", envelope.Type,
		"svix_id", svixID,
	)

	// Step 5: Route by event type
	ctx := r.Context()
	switch envelope.Type {
	case "organization.created", "organization.updated":
		h.handleOrgEvent(ctx, w, envelope)
	case "organizationMembership.created":
		h.handleMembershipEvent(ctx, w, envelope)
	default:
		// Silently ignore unhandled events
		slog.Debug("webhook/clerk: ignoring event", "type", envelope.Type)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ignored"}`))
	}
}

// handleOrgEvent upserts an organization by external_id.
func (h *ClerkWebhookHandler) handleOrgEvent(ctx context.Context, w http.ResponseWriter, envelope clerkWebhookEnvelope) {
	var orgData clerkOrgData
	if err := json.Unmarshal(envelope.Data, &orgData); err != nil {
		slog.Warn("webhook/clerk: failed to parse org data", "error", err)
		response.JSONError(w, http.StatusBadRequest, "Invalid org data")
		return
	}

	if orgData.ID == "" {
		slog.Warn("webhook/clerk: org event missing ID")
		response.JSONError(w, http.StatusBadRequest, "Missing organization ID")
		return
	}

	name := orgData.Name
	if name == "" {
		name = orgData.Slug
	}

	_, err := h.db.Exec(ctx,
		`INSERT INTO organizations (id, name, external_id, created_at)
		 VALUES (gen_random_uuid(), $1, $2, NOW())
		 ON CONFLICT (external_id)
		 DO UPDATE SET name = EXCLUDED.name`,
		name, orgData.ID,
	)
	if err != nil {
		slog.Error("webhook/clerk: failed to upsert organization",
			"error", err,
			"external_id", orgData.ID,
		)
		response.JSONError(w, http.StatusInternalServerError, "Failed to sync organization")
		return
	}

	slog.Info("webhook/clerk: organization synced",
		"external_id", orgData.ID,
		"name", name,
		"event", envelope.Type,
	)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"synced"}`))
}

// handleMembershipEvent upserts a user when they join an organization.
// This approach avoids the org_id NOT NULL constraint issue that occurs
// when handling user.created (which fires before org membership is assigned).
func (h *ClerkWebhookHandler) handleMembershipEvent(ctx context.Context, w http.ResponseWriter, envelope clerkWebhookEnvelope) {
	var memberData clerkMembershipData
	if err := json.Unmarshal(envelope.Data, &memberData); err != nil {
		slog.Warn("webhook/clerk: failed to parse membership data", "error", err)
		response.JSONError(w, http.StatusBadRequest, "Invalid membership data")
		return
	}

	if memberData.PublicUserData.UserID == "" || memberData.Organization.ID == "" {
		slog.Warn("webhook/clerk: membership event missing user or org ID")
		response.JSONError(w, http.StatusBadRequest, "Missing user or organization ID")
		return
	}

	// Map Clerk role to internal role
	role := "Builder"
	if memberData.Role == "org:admin" {
		role = "Admin"
	}

	name := memberData.PublicUserData.FirstName
	if memberData.PublicUserData.LastName != "" {
		name += " " + memberData.PublicUserData.LastName
	}
	if name == "" {
		name = memberData.PublicUserData.Identifier
	}

	email := memberData.PublicUserData.Identifier

	// Upsert user with org_id resolved from external_id
	_, err := h.db.Exec(ctx,
		`INSERT INTO users (id, org_id, email, name, role, external_id, created_at)
		 VALUES (
			gen_random_uuid(),
			(SELECT id FROM organizations WHERE external_id = $1),
			$2, $3, $4, $5, NOW()
		 )
		 ON CONFLICT (external_id)
		 DO UPDATE SET
			org_id = (SELECT id FROM organizations WHERE external_id = $1),
			name = EXCLUDED.name,
			role = EXCLUDED.role`,
		memberData.Organization.ID,
		email, name, role, memberData.PublicUserData.UserID,
	)
	if err != nil {
		slog.Error("webhook/clerk: failed to upsert user",
			"error", err,
			"external_id", memberData.PublicUserData.UserID,
			"org_external_id", memberData.Organization.ID,
		)
		response.JSONError(w, http.StatusInternalServerError, "Failed to sync user")
		return
	}

	slog.Info("webhook/clerk: user membership synced",
		"user_external_id", memberData.PublicUserData.UserID,
		"org_external_id", memberData.Organization.ID,
		"role", role,
	)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"synced"}`))
}

// verifySvixSignature validates the Svix HMAC-SHA256 signature.
// Svix signs: "{msgId}.{timestamp}.{body}" using the webhook secret.
// The signature header may contain multiple signatures separated by spaces.
func (h *ClerkWebhookHandler) verifySvixSignature(payload []byte, msgID, timestampStr, signatureHeader string) bool {
	if msgID == "" || timestampStr == "" || signatureHeader == "" {
		return false
	}

	// Validate timestamp to prevent replay attacks
	ts, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false
	}
	elapsed := time.Since(time.Unix(ts, 0))
	if elapsed < 0 {
		elapsed = -elapsed
	}
	if elapsed > clerkTimestampTolerance {
		slog.Warn("webhook/clerk: timestamp outside tolerance",
			"elapsed", elapsed,
			"tolerance", clerkTimestampTolerance,
		)
		return false
	}

	// Decode the webhook secret (Svix secrets are base64-encoded with "whsec_" prefix)
	secretStr := strings.TrimPrefix(h.webhookSecret, "whsec_")
	secretBytes, err := base64.StdEncoding.DecodeString(secretStr)
	if err != nil {
		slog.Warn("webhook/clerk: failed to decode webhook secret", "error", err)
		return false
	}

	// Compute expected signature: HMAC-SHA256("{msgId}.{timestamp}.{body}")
	toSign := fmt.Sprintf("%s.%s.%s", msgID, timestampStr, string(payload))
	mac := hmac.New(sha256.New, secretBytes)
	mac.Write([]byte(toSign))
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Check all signatures in the header (space-separated, each prefixed with "v1,")
	signatures := strings.Split(signatureHeader, " ")
	for _, sig := range signatures {
		parts := strings.SplitN(sig, ",", 2)
		if len(parts) != 2 {
			continue
		}
		version := parts[0]
		sigValue := parts[1]

		if version != "v1" {
			continue
		}

		// Constant-time comparison
		if hmac.Equal([]byte(sigValue), []byte(expected)) {
			return true
		}
	}

	return false
}

// --- Clerk Webhook Payload Types ---

// clerkWebhookEnvelope is the top-level structure of a Clerk webhook event.
type clerkWebhookEnvelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// clerkOrgData is the data payload for organization events.
type clerkOrgData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// clerkMembershipData is the data payload for organizationMembership events.
type clerkMembershipData struct {
	Role           string              `json:"role"`
	Organization   clerkMembershipOrg  `json:"organization"`
	PublicUserData clerkPublicUserData `json:"public_user_data"`
}

type clerkMembershipOrg struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type clerkPublicUserData struct {
	UserID     string `json:"user_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Identifier string `json:"identifier"`
}
