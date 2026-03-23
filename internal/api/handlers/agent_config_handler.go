package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/httputil"
	"github.com/google/uuid"
)

// AgentConfigHandler handles agent settings endpoints.
type AgentConfigHandler struct {
	svc *service.AgentConfigService
}

// NewAgentConfigHandler creates an AgentConfigHandler.
func NewAgentConfigHandler(svc *service.AgentConfigService) *AgentConfigHandler {
	return &AgentConfigHandler{svc: svc}
}

// GetAgentSettings handles GET /api/v1/org/settings/agents.
func (h *AgentConfigHandler) GetAgentSettings(w http.ResponseWriter, r *http.Request) {
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

	settings, err := h.svc.GetAgentConfig(ctx, orgID)
	if err != nil {
		slog.Error("agent config: failed to get", "error", err, "org_id", orgID)
		http.Error(w, "Failed to load agent settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(settings)
}

// UpdateAgentSettings handles PUT /api/v1/org/settings/agents.
func (h *AgentConfigHandler) UpdateAgentSettings(w http.ResponseWriter, r *http.Request) {
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

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user", http.StatusInternalServerError)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var settings config.AgentSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.svc.UpdateAgentConfig(ctx, orgID, userID, &settings); err != nil {
		slog.Error("agent config: failed to update", "error", err, "org_id", orgID)
		http.Error(w, "Failed to save agent settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(settings)
}
