package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/colton/futurebuild/internal/api/response"
	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// PolicyHandler handles autopilot policy API endpoints.
type PolicyHandler struct {
	policyEngine *service.PolicyEngine
}

// NewPolicyHandler creates a new policy handler.
func NewPolicyHandler(policyEngine *service.PolicyEngine) *PolicyHandler {
	return &PolicyHandler{policyEngine: policyEngine}
}

// ListPolicies returns all policies for the user's org.
// GET /api/v1/org/policies
func (h *PolicyHandler) ListPolicies() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		policies, err := h.policyEngine.ListPolicies(r.Context(), orgID)
		if err != nil {
			response.JSONError(w, http.StatusInternalServerError, "failed to list policies")
			return
		}
		if policies == nil {
			policies = []service.AutopilotPolicy{}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": policies})
	}
}

// UpsertPolicy creates or updates a policy for an action type.
// PUT /api/v1/org/policies/{actionType}
func (h *PolicyHandler) UpsertPolicy() http.HandlerFunc {
	type request struct {
		AutoApprove     bool  `json:"auto_approve"`
		MaxCostCents    int64 `json:"max_cost_cents"`
		CooldownMinutes int   `json:"cooldown_minutes"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
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

		actionType := chi.URLParam(r, "actionType")
		if actionType == "" {
			response.JSONError(w, http.StatusBadRequest, "action_type is required")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.JSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.MaxCostCents < 0 {
			response.JSONError(w, http.StatusBadRequest, "max_cost_cents must be non-negative")
			return
		}
		if req.CooldownMinutes < 0 {
			response.JSONError(w, http.StatusBadRequest, "cooldown_minutes must be non-negative")
			return
		}

		policy, err := h.policyEngine.UpsertPolicy(r.Context(), orgID, actionType, req.AutoApprove, req.MaxCostCents, req.CooldownMinutes)
		if err != nil {
			response.JSONError(w, http.StatusInternalServerError, "failed to update policy")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": policy})
	}
}
