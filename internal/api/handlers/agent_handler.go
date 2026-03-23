package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/service"
	"github.com/google/uuid"
)

// AgentHandler handles agent-related API endpoints.
type AgentHandler struct {
	agentActionService *service.AgentActionService
}

// NewAgentHandler creates a new AgentHandler.
func NewAgentHandler(aas *service.AgentActionService) *AgentHandler {
	return &AgentHandler{agentActionService: aas}
}

// ListPendingActions handles GET /api/v1/agents/pending.
// Returns all pending agent actions for the caller's organization.
func (h *AgentHandler) ListPendingActions(w http.ResponseWriter, r *http.Request) {
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

	actions, err := h.agentActionService.ListPending(ctx, orgID)
	if err != nil {
		slog.Error("agents: list pending failed", "error", err, "org_id", orgID)
		http.Error(w, "Failed to list pending actions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"pending_actions": actions,
		"count":           len(actions),
	})
}
