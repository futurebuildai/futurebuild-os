package handlers

import (
	"encoding/json"
	"log/slog"
	"math"
	"net/http"

	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/httputil"
	"github.com/google/uuid"
)

// ConfigHandler handles physics configuration endpoints.
// See STEP_87_CONFIG_PERSISTENCE.md Section 2
type ConfigHandler struct {
	configService service.ConfigServicer
}

// NewConfigHandler creates a new ConfigHandler.
func NewConfigHandler(cs service.ConfigServicer) *ConfigHandler {
	return &ConfigHandler{configService: cs}
}

// PhysicsConfigResponse is the API response for physics config.
type PhysicsConfigResponse struct {
	SpeedMultiplier float64 `json:"speed_multiplier"`
	WorkDays        []int   `json:"work_days"`
}

// UpdatePhysicsRequest is the request body for updating physics config.
type UpdatePhysicsRequest struct {
	SpeedMultiplier float64 `json:"speed_multiplier"`
	WorkDays        []int   `json:"work_days"`
}

// GetPhysics handles GET /api/v1/org/settings/physics.
// Returns the current organization's physics configuration.
func (h *ConfigHandler) GetPhysics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		slog.Warn("config: unauthorized - no claims in context", "error", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		slog.Error("config: invalid org_id in claims", "error", err)
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
		return
	}

	cfg, err := h.configService.GetConfig(ctx, orgID)
	if err != nil {
		slog.Error("config: failed to get config", "error", err, "org_id", orgID)
		http.Error(w, "Failed to load settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(PhysicsConfigResponse{
		SpeedMultiplier: cfg.SpeedMultiplier,
		WorkDays:        cfg.WorkDays,
	})
}

// UpdatePhysics handles PUT /api/v1/org/settings/physics.
// Updates the current organization's physics configuration.
func (h *ConfigHandler) UpdatePhysics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		slog.Warn("config: unauthorized - no claims in context", "error", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		slog.Error("config: invalid org_id in claims", "error", err)
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
		return
	}

	// L7: Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var req UpdatePhysicsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("config: invalid request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// H-1: Guard against NaN/Infinity from malformed JSON
	if math.IsNaN(req.SpeedMultiplier) || math.IsInf(req.SpeedMultiplier, 0) {
		http.Error(w, "speed_multiplier must be a finite number", http.StatusBadRequest)
		return
	}

	// C-2: Range aligned with frontend slider (0.5-1.5, not 2.0)
	if req.SpeedMultiplier < 0.5 || req.SpeedMultiplier > 1.5 {
		http.Error(w, "speed_multiplier must be between 0.5 and 1.5", http.StatusBadRequest)
		return
	}

	// Round to 1 decimal place to match frontend precision
	req.SpeedMultiplier = math.Round(req.SpeedMultiplier*10) / 10

	// Validate work_days: must be 1-7 values, each between 1-7
	if len(req.WorkDays) == 0 || len(req.WorkDays) > 7 {
		http.Error(w, "work_days must contain 1-7 day values", http.StatusBadRequest)
		return
	}
	seen := make(map[int]bool, 7)
	for _, d := range req.WorkDays {
		if d < 1 || d > 7 {
			http.Error(w, "work_days values must be between 1 (Mon) and 7 (Sun)", http.StatusBadRequest)
			return
		}
		if seen[d] {
			http.Error(w, "work_days must not contain duplicates", http.StatusBadRequest)
			return
		}
		seen[d] = true
	}

	cfg, err := h.configService.UpdateConfig(ctx, orgID, req.SpeedMultiplier, req.WorkDays)
	if err != nil {
		slog.Error("config: failed to update config", "error", err, "org_id", orgID)
		http.Error(w, "Failed to save settings", http.StatusInternalServerError)
		return
	}

	// H-5: Service already logs — no duplicate log here.

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(PhysicsConfigResponse{
		SpeedMultiplier: cfg.SpeedMultiplier,
		WorkDays:        cfg.WorkDays,
	})
}
