package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/colton/futurebuild/internal/api/response"
	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// BrainSettingsHandler handles FB-Brain connection settings and A2A logging endpoints.
// See FRONTEND_SCOPE.md Section 15.1
type BrainSettingsHandler struct {
	svc    *service.BrainConnectionService
	a2aSvc service.A2AServicer
}

// NewBrainSettingsHandler creates a BrainSettingsHandler.
func NewBrainSettingsHandler(svc *service.BrainConnectionService) *BrainSettingsHandler {
	return &BrainSettingsHandler{svc: svc}
}

// WithA2AService injects the A2A service for agent/log endpoints.
func (h *BrainSettingsHandler) WithA2AService(a2aSvc service.A2AServicer) {
	h.a2aSvc = a2aSvc
}

// GetBrainConnection handles GET /api/v1/org/settings/brain.
func (h *BrainSettingsHandler) GetBrainConnection(w http.ResponseWriter, r *http.Request) {
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

	conn, err := h.svc.GetConnection(ctx, orgID)
	if err != nil {
		slog.Error("brain settings: failed to get", "error", err, "org_id", orgID)
		http.Error(w, "Failed to load brain settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(conn)
}

// updateBrainRequest is the request body for updating brain connection.
type updateBrainRequest struct {
	BrainURL string `json:"brain_url"`
}

// UpdateBrainConnection handles PUT /api/v1/org/settings/brain.
func (h *BrainSettingsHandler) UpdateBrainConnection(w http.ResponseWriter, r *http.Request) {
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
	var req updateBrainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.svc.UpdateConnection(ctx, orgID, req.BrainURL); err != nil {
		slog.Error("brain settings: failed to update", "error", err, "org_id", orgID)
		http.Error(w, "Failed to save brain settings", http.StatusInternalServerError)
		return
	}

	conn, err := h.svc.GetConnection(ctx, orgID)
	if err != nil {
		http.Error(w, "Failed to reload settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(conn)
}

// RegenerateKey handles POST /api/v1/org/settings/brain/regenerate-key.
func (h *BrainSettingsHandler) RegenerateKey(w http.ResponseWriter, r *http.Request) {
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

	key, err := h.svc.RegenerateKey(ctx, orgID)
	if err != nil {
		slog.Error("brain settings: failed to regenerate key", "error", err, "org_id", orgID)
		http.Error(w, "Failed to regenerate key", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"integration_key": key})
}

// GetActiveAgents handles GET /api/v1/org/settings/brain/agents.
func (h *BrainSettingsHandler) GetActiveAgents(w http.ResponseWriter, r *http.Request) {
	if h.a2aSvc == nil {
		response.JSONError(w, http.StatusServiceUnavailable, "A2A service not configured")
		return
	}

	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		response.JSONError(w, http.StatusUnauthorized, "missing auth context")
		return
	}

	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		response.JSONError(w, http.StatusUnauthorized, "invalid org context")
		return
	}

	agents, err := h.a2aSvc.GetActiveAgents(r.Context(), orgID)
	if err != nil {
		slog.Error("brain: failed to get active agents", "error", err, "org_id", orgID)
		response.JSONError(w, http.StatusInternalServerError, "failed to get active agents")
		return
	}
	if agents == nil {
		agents = []models.ActiveAgentConnection{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": agents})
}

// PauseAgent handles POST /api/v1/org/settings/brain/agents/{id}/pause.
func (h *BrainSettingsHandler) PauseAgent(w http.ResponseWriter, r *http.Request) {
	if h.a2aSvc == nil {
		response.JSONError(w, http.StatusServiceUnavailable, "A2A service not configured")
		return
	}

	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		response.JSONError(w, http.StatusUnauthorized, "missing auth context")
		return
	}

	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		response.JSONError(w, http.StatusUnauthorized, "invalid org context")
		return
	}

	agentID, err := uuid.Parse(chi.URLParam(r, "agentId"))
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	if err := h.a2aSvc.PauseAgent(r.Context(), agentID, orgID); err != nil {
		slog.Error("brain: failed to pause agent", "error", err, "agent_id", agentID)
		response.JSONError(w, http.StatusInternalServerError, "failed to pause agent")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "paused"})
}

// ResumeAgent handles POST /api/v1/org/settings/brain/agents/{id}/resume.
func (h *BrainSettingsHandler) ResumeAgent(w http.ResponseWriter, r *http.Request) {
	if h.a2aSvc == nil {
		response.JSONError(w, http.StatusServiceUnavailable, "A2A service not configured")
		return
	}

	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		response.JSONError(w, http.StatusUnauthorized, "missing auth context")
		return
	}

	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		response.JSONError(w, http.StatusUnauthorized, "invalid org context")
		return
	}

	agentID, err := uuid.Parse(chi.URLParam(r, "agentId"))
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	if err := h.a2aSvc.ResumeAgent(r.Context(), agentID, orgID); err != nil {
		slog.Error("brain: failed to resume agent", "error", err, "agent_id", agentID)
		response.JSONError(w, http.StatusInternalServerError, "failed to resume agent")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "active"})
}

// GetExecutionLogs handles GET /api/v1/org/settings/brain/logs?limit=.
func (h *BrainSettingsHandler) GetExecutionLogs(w http.ResponseWriter, r *http.Request) {
	if h.a2aSvc == nil {
		response.JSONError(w, http.StatusServiceUnavailable, "A2A service not configured")
		return
	}

	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		response.JSONError(w, http.StatusUnauthorized, "missing auth context")
		return
	}

	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		response.JSONError(w, http.StatusUnauthorized, "invalid org context")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	logs, err := h.a2aSvc.GetExecutionLogs(r.Context(), orgID, limit)
	if err != nil {
		slog.Error("brain: failed to get execution logs", "error", err, "org_id", orgID)
		response.JSONError(w, http.StatusInternalServerError, "failed to get execution logs")
		return
	}
	if logs == nil {
		logs = []models.A2AExecutionLog{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": logs})
}
