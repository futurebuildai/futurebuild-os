package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/colton/futurebuild/internal/readiness"
)

// ReadinessHandler serves the deep integration readiness endpoint.
type ReadinessHandler struct {
	service *readiness.Service
	env     string
}

// NewReadinessHandler creates a handler backed by the given readiness service.
func NewReadinessHandler(svc *readiness.Service, env string) *ReadinessHandler {
	return &ReadinessHandler{service: svc, env: env}
}

// HandleReadiness runs all integration probes and returns the aggregated report.
// Returns 200 for healthy/degraded and 503 for failed.
func (h *ReadinessHandler) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	report := h.service.Run(r.Context(), h.env)

	w.Header().Set("Content-Type", "application/json")
	if report.Status == readiness.StatusFailed {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_ = json.NewEncoder(w).Encode(report)
}
