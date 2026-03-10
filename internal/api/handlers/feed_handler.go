package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/httputil"
	"github.com/google/uuid"
)

// FeedHandler handles portfolio feed endpoints.
// See FRONTEND_V2_SPEC.md §5.1, §5.2
type FeedHandler struct {
	feedService        *service.FeedService
	integrationClient  *IntegrationClient
}

// NewFeedHandler creates a new FeedHandler.
func NewFeedHandler(fs *service.FeedService) *FeedHandler {
	return &FeedHandler{
		feedService:       fs,
		integrationClient: NewIntegrationClient(),
	}
}

// GetFeed handles GET /api/v1/portfolio/feed.
// Returns the portfolio feed with greeting, summary, and cards.
func (h *FeedHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		slog.Warn("feed: unauthorized - no claims in context", "error", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Validate org_id is present and valid
	if claims.OrgID == "" {
		slog.Warn("feed: user has no organization",
			"user_id", claims.UserID,
			"email", claims.Email)
		http.Error(w, "User not associated with an organization", http.StatusForbidden)
		return
	}

	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		slog.Error("feed: invalid org_id format in claims",
			"error", err,
			"org_id_raw", claims.OrgID,
			"user_id", claims.UserID)
		http.Error(w, "Invalid organization ID format", http.StatusInternalServerError)
		return
	}

	// Optional project filter
	var projectFilter *uuid.UUID
	if pidStr := r.URL.Query().Get("project_id"); pidStr != "" {
		pid, err := uuid.Parse(pidStr)
		if err != nil {
			http.Error(w, "Invalid project_id parameter", http.StatusBadRequest)
			return
		}
		projectFilter = &pid
	}

	feed, err := h.feedService.GetFeed(ctx, orgID, projectFilter)
	if err != nil {
		slog.Error("feed: failed to get feed", "error", err, "org_id", orgID)
		http.Error(w, "Failed to load feed", http.StatusInternalServerError)
		return
	}

	// Get active projects for project pills
	projects, err := h.feedService.ListActiveProjectsForOrg(ctx, orgID)
	if err != nil {
		slog.Error("feed: failed to list projects", "error", err, "org_id", orgID)
		http.Error(w, "Failed to load projects", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(PortfolioFeedResponse{
		Greeting: feed.Greeting,
		Summary:  feed.Summary,
		Cards:    feed.Cards,
		Projects: projects,
	})
}

// ActionResponse is the structured response from ExecuteAction.
// The frontend uses `effect` to determine post-action behavior:
//   - "dismiss": remove the card from the feed list
//   - "navigate": client-side route to `navigate_to`
//   - "none": no visual change (informational)
type ActionResponse struct {
	Success    bool                   `json:"success"`
	Effect     string                 `json:"effect"`
	Message    string                 `json:"message,omitempty"`
	NavigateTo string                 `json:"navigate_to,omitempty"`
	Payload    map[string]interface{} `json:"payload,omitempty"`
}

// formatSnoozeLabel returns a human-readable snooze duration.
func formatSnoozeLabel(hours int) string {
	if hours < 24 {
		return fmt.Sprintf("%dh", hours)
	}
	days := hours / 24
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}

// PortfolioFeedResponse wraps the feed response with project list for pills.
type PortfolioFeedResponse struct {
	Greeting string                        `json:"greeting"`
	Summary  service.PortfolioSummary      `json:"summary"`
	Cards    []models.FeedCard             `json:"cards"`
	Projects []models.Project              `json:"projects"`
}

// DismissCard handles POST /api/v1/portfolio/feed/dismiss.
func (h *FeedHandler) DismissCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var req struct {
		CardID string `json:"card_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cardID, err := uuid.Parse(req.CardID)
	if err != nil {
		http.Error(w, "Invalid card_id", http.StatusBadRequest)
		return
	}

	if err := h.feedService.DismissCard(ctx, orgID, cardID); err != nil {
		slog.Error("feed: dismiss failed", "error", err, "card_id", cardID)
		http.Error(w, "Failed to dismiss card", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// SnoozeCard handles POST /api/v1/portfolio/feed/snooze.
func (h *FeedHandler) SnoozeCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var req struct {
		CardID string `json:"card_id"`
		Hours  int    `json:"hours"` // Snooze duration in hours
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cardID, err := uuid.Parse(req.CardID)
	if err != nil {
		http.Error(w, "Invalid card_id", http.StatusBadRequest)
		return
	}

	if req.Hours < 1 || req.Hours > 168 { // Max 7 days
		http.Error(w, "Hours must be between 1 and 168", http.StatusBadRequest)
		return
	}

	until := time.Now().Add(time.Duration(req.Hours) * time.Hour)
	if err := h.feedService.SnoozeCard(ctx, orgID, cardID, until); err != nil {
		slog.Error("feed: snooze failed", "error", err, "card_id", cardID)
		http.Error(w, "Failed to snooze card", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// ExecuteAction handles POST /api/v1/portfolio/feed/action.
// This is a command dispatcher that routes to existing service methods.
func (h *FeedHandler) ExecuteAction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var req struct {
		CardID   string                 `json:"card_id"`
		ActionID string                 `json:"action_id"`
		Payload  map[string]interface{} `json:"payload,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.CardID == "" || req.ActionID == "" {
		http.Error(w, "card_id and action_id are required", http.StatusBadRequest)
		return
	}

	cardID, err := uuid.Parse(req.CardID)
	if err != nil {
		http.Error(w, "Invalid card_id", http.StatusBadRequest)
		return
	}

	resp := ActionResponse{Success: true, Effect: "none"}

	switch req.ActionID {
	case "dismiss":
		if err := h.feedService.DismissCard(ctx, orgID, cardID); err != nil {
			http.Error(w, "Failed to dismiss card", http.StatusInternalServerError)
			return
		}
		resp.Effect = "dismiss"
		resp.Message = "Card dismissed"

	case "snooze":
		hours := 24
		if h, ok := req.Payload["hours"].(float64); ok && h > 0 && h <= 168 {
			hours = int(h)
		}
		until := time.Now().Add(time.Duration(hours) * time.Hour)
		if err := h.feedService.SnoozeCard(ctx, orgID, cardID, until); err != nil {
			http.Error(w, "Failed to snooze card", http.StatusInternalServerError)
			return
		}
		resp.Effect = "dismiss"
		resp.Message = "Snoozed for " + formatSnoozeLabel(hours)

	case "order_now":
		if err := h.feedService.MarkProcurementOrdered(ctx, orgID, cardID); err != nil {
			slog.Error("feed: mark ordered failed", "error", err, "card_id", cardID)
			http.Error(w, "Failed to mark as ordered", http.StatusInternalServerError)
			return
		}
		resp.Effect = "dismiss"
		resp.Message = "Item marked as ordered"

	case "view_briefing", "view_details":
		// Look up project from card for navigation
		card, err := h.feedService.GetCardByID(ctx, orgID, cardID)
		if err != nil {
			http.Error(w, "Card not found", http.StatusNotFound)
			return
		}
		resp.Effect = "navigate"
		resp.NavigateTo = "/project/" + card.ProjectID.String()

	case "view_schedule":
		card, err := h.feedService.GetCardByID(ctx, orgID, cardID)
		if err != nil {
			http.Error(w, "Card not found", http.StatusNotFound)
			return
		}
		resp.Effect = "navigate"
		resp.NavigateTo = "/project/" + card.ProjectID.String() + "/schedule"

	case "resend":
		// Re-sending sub confirmation is complex (needs SubLiaisonAgent).
		// For now, snooze 1h so the agent picks it up on next scan.
		until := time.Now().Add(1 * time.Hour)
		if err := h.feedService.SnoozeCard(ctx, orgID, cardID, until); err != nil {
			http.Error(w, "Failed to resend", http.StatusInternalServerError)
			return
		}
		resp.Effect = "dismiss"
		resp.Message = "Confirmation will be resent shortly"

	case "call_sub":
		// Frontend-only action — return the task context for dialer
		card, err := h.feedService.GetCardByID(ctx, orgID, cardID)
		if err != nil {
			http.Error(w, "Card not found", http.StatusNotFound)
			return
		}
		resp.Effect = "none"
		resp.Message = "Contact information loaded"
		if card.TaskID != nil {
			resp.Payload = map[string]interface{}{"task_id": card.TaskID.String()}
		}

	case "add_contacts":
		// Navigate to contact directory — card stays visible
		resp.Effect = "navigate"
		resp.NavigateTo = "/contacts"

	case "assign_contact":
		// Contact was assigned inline — dismiss the setup_contacts card
		if err := h.feedService.DismissCard(ctx, orgID, cardID); err != nil {
			http.Error(w, "Failed to dismiss card", http.StatusInternalServerError)
			return
		}
		resp.Effect = "dismiss"
		resp.Message = "Contact assigned"

	// ── Integration actions (FB-Brain cross-system flows) ──

	case "request_material_quote":
		card, err := h.feedService.GetCardByID(ctx, orgID, cardID)
		if err != nil {
			http.Error(w, "Card not found", http.StatusNotFound)
			return
		}
		flowResp, err := h.integrationClient.StartMaterialsFlow(orgID.String(), card.ProjectID.String(), cardID.String())
		if err != nil {
			slog.Error("feed: start materials flow failed", "error", err)
			http.Error(w, "Failed to request material quote: "+err.Error(), http.StatusInternalServerError)
			return
		}
		_ = h.feedService.DismissCard(ctx, orgID, cardID)
		resp.Effect = "dismiss"
		resp.Message = flowResp.Message

	case "approve_material_quote":
		rfqID := ""
		if v, ok := req.Payload["rfq_id"].(string); ok {
			rfqID = v
		}
		// Also check card engine_data for rfq_id
		if rfqID == "" {
			card, err := h.feedService.GetCardByID(ctx, orgID, cardID)
			if err == nil && card.EngineData != nil {
				var ed map[string]interface{}
				json.Unmarshal(card.EngineData, &ed)
				if v, ok := ed["rfq_id"].(string); ok {
					rfqID = v
				}
			}
		}
		if rfqID == "" {
			http.Error(w, "rfq_id required", http.StatusBadRequest)
			return
		}
		flowResp, err := h.integrationClient.ApproveMaterialsQuote(rfqID)
		if err != nil {
			slog.Error("feed: approve materials quote failed", "error", err)
			http.Error(w, "Failed to approve quote: "+err.Error(), http.StatusInternalServerError)
			return
		}
		_ = h.feedService.DismissCard(ctx, orgID, cardID)
		resp.Effect = "dismiss"
		resp.Message = flowResp.Message

	case "send_labor_rfq":
		card, err := h.feedService.GetCardByID(ctx, orgID, cardID)
		if err != nil {
			http.Error(w, "Card not found", http.StatusNotFound)
			return
		}
		flowResp, err := h.integrationClient.StartLaborFlow(orgID.String(), card.ProjectID.String(), cardID.String())
		if err != nil {
			slog.Error("feed: start labor flow failed", "error", err)
			http.Error(w, "Failed to send labor RFQ: "+err.Error(), http.StatusInternalServerError)
			return
		}
		_ = h.feedService.DismissCard(ctx, orgID, cardID)
		resp.Effect = "dismiss"
		resp.Message = flowResp.Message

	case "approve_labor_bid":
		rfqID := ""
		if v, ok := req.Payload["rfq_id"].(string); ok {
			rfqID = v
		}
		if rfqID == "" {
			card, err := h.feedService.GetCardByID(ctx, orgID, cardID)
			if err == nil && card.EngineData != nil {
				var ed map[string]interface{}
				json.Unmarshal(card.EngineData, &ed)
				if v, ok := ed["rfq_id"].(string); ok {
					rfqID = v
				}
			}
		}
		if rfqID == "" {
			http.Error(w, "rfq_id required", http.StatusBadRequest)
			return
		}
		flowResp, err := h.integrationClient.ApproveLaborBid(rfqID)
		if err != nil {
			slog.Error("feed: approve labor bid failed", "error", err)
			http.Error(w, "Failed to approve bid: "+err.Error(), http.StatusInternalServerError)
			return
		}
		_ = h.feedService.DismissCard(ctx, orgID, cardID)
		resp.Effect = "dismiss"
		resp.Message = flowResp.Message

	case "confirm_delivery":
		rfqID := ""
		if v, ok := req.Payload["rfq_id"].(string); ok {
			rfqID = v
		}
		if rfqID == "" {
			// No rfq needed for delivery confirm — just dismiss
			_ = h.feedService.DismissCard(ctx, orgID, cardID)
			resp.Effect = "dismiss"
			resp.Message = "Delivery confirmed"
			break
		}
		flowResp, err := h.integrationClient.ConfirmDelivery(rfqID)
		if err != nil {
			slog.Error("feed: confirm delivery failed", "error", err)
		}
		_ = h.feedService.DismissCard(ctx, orgID, cardID)
		resp.Effect = "dismiss"
		if flowResp != nil {
			resp.Message = flowResp.Message
		} else {
			resp.Message = "Delivery confirmed"
		}

	default:
		slog.Info("feed: unhandled action", "action_id", req.ActionID, "card_id", req.CardID, "org_id", orgID)
		http.Error(w, "Unknown action: "+req.ActionID, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// StreamFeed handles GET /api/v1/portfolio/feed/stream.
// Implements Server-Sent Events (SSE) for live feed card updates.
// See FRONTEND_V2_SPEC.md §6.5 — uses PostgreSQL LISTEN/NOTIFY under the hood.
func (h *FeedHandler) StreamFeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
		return
	}

	// Verify SSE support (flusher interface)
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Subscribe to feed card notifications for this org
	ch, unsub, err := h.feedService.SubscribeFeedChanges(ctx, orgID)
	if err != nil {
		slog.Error("feed: failed to subscribe to changes", "error", err, "org_id", orgID)
		http.Error(w, "Failed to open stream", http.StatusInternalServerError)
		return
	}
	defer unsub()

	// Send initial keepalive comment to confirm connection
	_, _ = fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	// Keepalive ticker to prevent proxy timeouts
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		case event, ok := <-ch:
			if !ok {
				return // Channel closed
			}
			data, err := json.Marshal(event.Payload)
			if err != nil {
				slog.Error("feed: failed to marshal SSE event", "error", err)
				continue
			}
			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
			flusher.Flush()
		}
	}
}
