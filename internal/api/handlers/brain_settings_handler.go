package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/httputil"
	"github.com/google/uuid"
)

// BrainSettingsHandler handles FB-Brain connection settings endpoints.
type BrainSettingsHandler struct {
	svc *service.BrainConnectionService
}

// NewBrainSettingsHandler creates a BrainSettingsHandler.
func NewBrainSettingsHandler(svc *service.BrainConnectionService) *BrainSettingsHandler {
	return &BrainSettingsHandler{svc: svc}
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
