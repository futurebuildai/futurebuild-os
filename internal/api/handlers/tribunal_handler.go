package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/colton/futurebuild/internal/api/response"
	"github.com/colton/futurebuild/internal/futureshade/tribunal"
	"github.com/colton/futurebuild/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// TribunalHandler handles Tribunal decision endpoints.
// See SHADOW_VIEWER_specs.md Section 3.1
type TribunalHandler struct {
	repo *tribunal.Repository
}

// NewTribunalHandler creates a new TribunalHandler.
func NewTribunalHandler(repo *tribunal.Repository) *TribunalHandler {
	return &TribunalHandler{repo: repo}
}

// ListDecisions returns paginated tribunal decisions with optional filtering.
// GET /api/v1/tribunal/decisions
// Query params: limit, offset, status, model, start_date, end_date, search
func (h *TribunalHandler) ListDecisions(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filter := tribunal.ListDecisionsFilter{
		Limit:  parseIntOrDefault(r.URL.Query().Get("limit"), 20),
		Offset: parseIntOrDefault(r.URL.Query().Get("offset"), 0),
		Status: tribunal.DecisionStatus(r.URL.Query().Get("status")),
		Model:  r.URL.Query().Get("model"),
		Search: r.URL.Query().Get("search"),
	}

	// Parse dates if provided
	if start := r.URL.Query().Get("start_date"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			filter.StartDate = &t
		}
	}
	if end := r.URL.Query().Get("end_date"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			filter.EndDate = &t
		}
	}

	// Audit log: Shadow Mode accessed
	claims, _ := middleware.GetClaims(r.Context())
	slog.Info("tribunal: decisions listed",
		"user_id", claims.UserID,
		"filter_status", filter.Status,
		"filter_model", filter.Model,
	)

	// Handle nil repo (Fail Open)
	if h.repo == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tribunal.ListDecisionsResponse{
			Decisions: []tribunal.DecisionSummary{},
			Total:     0,
			HasMore:   false,
		})
		return
	}

	result, err := h.repo.ListDecisions(r.Context(), filter)
	if err != nil {
		slog.Error("tribunal: list failed", "error", err)
		response.JSONError(w, http.StatusInternalServerError, "Failed to list decisions")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetDecision returns a single decision with its votes.
// GET /api/v1/tribunal/decisions/{id}
func (h *TribunalHandler) GetDecision(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "Invalid decision ID")
		return
	}

	// Audit log
	claims, _ := middleware.GetClaims(r.Context())
	slog.Info("tribunal: decision viewed",
		"user_id", claims.UserID,
		"decision_id", id,
	)

	// Handle nil repo (Fail Closed - return 503)
	if h.repo == nil {
		response.JSONError(w, http.StatusServiceUnavailable, "Tribunal service unavailable")
		return
	}

	decision, err := h.repo.GetDecision(r.Context(), id)
	if err != nil {
		if err == tribunal.ErrNotFound {
			response.JSONError(w, http.StatusNotFound, "Decision not found")
			return
		}
		slog.Error("tribunal: get failed", "error", err, "id", id)
		response.JSONError(w, http.StatusInternalServerError, "Failed to get decision")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decision)
}

// parseIntOrDefault parses an integer from string or returns default.
func parseIntOrDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
