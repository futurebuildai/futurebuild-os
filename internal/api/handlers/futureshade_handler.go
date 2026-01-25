package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/colton/futurebuild/internal/futureshade"
)

// FutureShadeHandler handles FutureShade-related HTTP endpoints.
// See FUTURESHADE_INIT_specs.md Section 3.2.
type FutureShadeHandler struct {
	service futureshade.Service
}

// NewFutureShadeHandler creates a new FutureShadeHandler.
// Accepts nil service for Fail Open behavior - will return disabled status.
func NewFutureShadeHandler(service futureshade.Service) *FutureShadeHandler {
	return &FutureShadeHandler{service: service}
}

// HealthResponse is the response payload for the health endpoint.
type HealthResponse struct {
	Status        string `json:"status"`
	TribunalCount int    `json:"tribunal_count"`
}

// HandleHealth returns the health status of the FutureShade service.
// GET /api/v1/futureshade/health
// Admin-only endpoint (middleware enforced).
func (h *FutureShadeHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	// Nil-safety check for Fail Open pattern
	if h.service == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(HealthResponse{
			Status:        "disabled",
			TribunalCount: 0,
		})
		return
	}

	err := h.service.Health()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(HealthResponse{
			Status:        "disabled",
			TribunalCount: 0,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HealthResponse{
		Status:        "active",
		TribunalCount: 0, // TODO: Step 65+ will return actual tribunal count
	})
}
