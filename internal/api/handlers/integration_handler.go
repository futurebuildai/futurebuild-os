package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/a2a"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IntegrationHandler receives webhooks from FB-Brain to create feed cards
// and assign contacts to project phases.
type IntegrationHandler struct {
	feedService *service.FeedService
	db          *pgxpool.Pool
	apiKey      string
}

func NewIntegrationHandler(fs *service.FeedService, db *pgxpool.Pool) *IntegrationHandler {
	apiKey := os.Getenv("INTEGRATION_API_KEY")
	if apiKey == "" {
		apiKey = "fb-brain-demo-key-2026"
	}
	return &IntegrationHandler{
		feedService: fs,
		db:          db,
		apiKey:      apiKey,
	}
}

// ValidateAPIKey checks the X-Integration-Key header (legacy, kept for reference).
func (h *IntegrationHandler) ValidateAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Integration-Key")
		if key != h.apiKey {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ValidateSignature is the preferred middleware for authenticating FB-Brain requests.
// It supports two paths:
//   - HMAC (preferred): X-Signature + X-Timestamp + X-Nonce headers verified via a2a.VerifySignature
//   - Legacy fallback: X-Integration-Key string match (deprecated, logs warning)
//
// Both paths are fail-closed: invalid credentials are rejected.
func (h *IntegrationHandler) ValidateSignature(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sig := r.Header.Get("X-Signature")
		ts := r.Header.Get("X-Timestamp")
		nonce := r.Header.Get("X-Nonce")

		// HMAC path: all three headers present
		if sig != "" && ts != "" && nonce != "" {
			// Buffer the body so downstream handlers can still read it
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, `{"error":"failed to read body"}`, http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			// Verify HMAC signature
			if !a2a.VerifySignature(h.apiKey, sig, ts, nonce, bodyBytes) {
				slog.Warn("integration: HMAC signature verification failed",
					"remote_addr", r.RemoteAddr)
				http.Error(w, `{"error":"invalid signature"}`, http.StatusUnauthorized)
				return
			}

			// Check timestamp freshness (300s = 5 minutes)
			tsInt, err := strconv.ParseInt(ts, 10, 64)
			if err != nil {
				http.Error(w, `{"error":"invalid timestamp"}`, http.StatusUnauthorized)
				return
			}
			drift := math.Abs(float64(time.Now().Unix() - tsInt))
			if drift > 300 {
				slog.Warn("integration: timestamp drift too large",
					"drift_seconds", drift, "remote_addr", r.RemoteAddr)
				http.Error(w, `{"error":"request expired"}`, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
			return
		}

		// Legacy fallback: plain API key match
		key := r.Header.Get("X-Integration-Key")
		if key == h.apiKey {
			slog.Warn("integration: legacy X-Integration-Key auth used — upgrade to HMAC signing",
				"remote_addr", r.RemoteAddr,
				"path", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}

		// Both paths failed
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
	})
}

// CreateFeedCardRequest is the inbound payload from FB-Brain.
type CreateFeedCardRequest struct {
	OrgID      string          `json:"org_id"`
	ProjectID  string          `json:"project_id"`
	CardType   string          `json:"card_type"`
	Headline   string          `json:"headline"`
	Body       string          `json:"body"`
	Priority   int             `json:"priority"`
	Horizon    string          `json:"horizon"`
	Actions    json.RawMessage `json:"actions"`
	EngineData json.RawMessage `json:"engine_data,omitempty"`
}

// CreateFeedCard handles POST /api/v1/integration/feed-card.
// FB-Brain calls this to inject integration feed cards into XUI.
func (h *IntegrationHandler) CreateFeedCard(w http.ResponseWriter, r *http.Request) {
	var req CreateFeedCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	orgID, err := uuid.Parse(req.OrgID)
	if err != nil {
		http.Error(w, `{"error":"invalid org_id"}`, http.StatusBadRequest)
		return
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		http.Error(w, `{"error":"invalid project_id"}`, http.StatusBadRequest)
		return
	}

	// Parse actions
	var actions []models.FeedCardAction
	if req.Actions != nil {
		json.Unmarshal(req.Actions, &actions)
	}

	card := &models.FeedCard{
		ID:         uuid.New(),
		OrgID:      orgID,
		ProjectID:  projectID,
		CardType:   models.FeedCardType(req.CardType),
		Priority:   req.Priority,
		Headline:   req.Headline,
		Body:       req.Body,
		Horizon:    models.FeedCardHorizon(req.Horizon),
		Actions:    actions,
		EngineData: req.EngineData,
	}

	if err := h.feedService.WriteCard(r.Context(), card); err != nil {
		slog.Error("integration: write card failed", "error", err)
		http.Error(w, `{"error":"failed to write card"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": card.ID.String()})
}

// AssignContactRequest is the payload for assigning a contact to a WBS phase.
type AssignContactRequest struct {
	ProjectID string `json:"project_id"`
	ContactID string `json:"contact_id"`
	WBSPhase  string `json:"wbs_phase"`
}

// AssignContact handles POST /api/v1/integration/assign-contact.
func (h *IntegrationHandler) AssignContact(w http.ResponseWriter, r *http.Request) {
	var req AssignContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		http.Error(w, `{"error":"invalid project_id"}`, http.StatusBadRequest)
		return
	}

	contactID, err := uuid.Parse(req.ContactID)
	if err != nil {
		http.Error(w, `{"error":"invalid contact_id"}`, http.StatusBadRequest)
		return
	}

	_, err = h.db.Exec(r.Context(), `
		INSERT INTO project_assignments (id, project_id, contact_id, wbs_phase_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING
	`, uuid.New(), projectID, contactID, req.WBSPhase)
	if err != nil {
		slog.Error("integration: assign contact failed", "error", err)
		http.Error(w, `{"error":"failed to assign contact"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
