package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/httputil"
	"github.com/google/uuid"
)

// FeedHandler handles portfolio feed endpoints.
// See FRONTEND_V2_SPEC.md §5.1, §5.2
type FeedHandler struct {
	feedService *service.FeedService
}

// NewFeedHandler creates a new FeedHandler.
func NewFeedHandler(fs *service.FeedService) *FeedHandler {
	return &FeedHandler{feedService: fs}
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

	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		slog.Error("feed: invalid org_id in claims", "error", err)
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
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

// PortfolioFeedResponse wraps the feed response with project list for pills.
type PortfolioFeedResponse struct {
	Greeting string                        `json:"greeting"`
	Summary  service.PortfolioSummary      `json:"summary"`
	Cards    interface{}                   `json:"cards"`
	Projects interface{}                   `json:"projects"`
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

	// For now, handle dismiss and snooze actions.
	// Additional action routing (order_now, confirm_start, etc.) will be added in Phase 3.
	switch req.ActionID {
	case "dismiss":
		cardID, err := uuid.Parse(req.CardID)
		if err != nil {
			http.Error(w, "Invalid card_id", http.StatusBadRequest)
			return
		}
		if err := h.feedService.DismissCard(ctx, orgID, cardID); err != nil {
			http.Error(w, "Failed to dismiss card", http.StatusInternalServerError)
			return
		}
	case "snooze":
		cardID, err := uuid.Parse(req.CardID)
		if err != nil {
			http.Error(w, "Invalid card_id", http.StatusBadRequest)
			return
		}
		// Default snooze: 24 hours
		until := time.Now().Add(24 * time.Hour)
		if err := h.feedService.SnoozeCard(ctx, orgID, cardID, until); err != nil {
			http.Error(w, "Failed to snooze card", http.StatusInternalServerError)
			return
		}
	default:
		// Phase 3: Route to appropriate service methods
		slog.Info("feed: unhandled action", "action_id", req.ActionID, "card_id", req.CardID, "org_id", orgID)
		http.Error(w, "Action not yet implemented", http.StatusNotImplemented)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Action executed",
	})
}
