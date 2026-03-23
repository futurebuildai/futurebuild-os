package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/colton/futurebuild/internal/service"
)

const maxPreviewBodySize = 1 << 20 // 1 MB

// PreviewHandler exposes schedule preview and scenario comparison endpoints.
type PreviewHandler struct {
	previewSvc *service.SchedulePreviewService
}

// NewPreviewHandler creates a new preview handler.
func NewPreviewHandler(previewSvc *service.SchedulePreviewService) *PreviewHandler {
	return &PreviewHandler{previewSvc: previewSvc}
}

// GeneratePreview handles POST /api/v1/schedule/preview
func (h *PreviewHandler) GeneratePreview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxPreviewBodySize)

		var req service.SchedulePreviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.StartDate == "" {
			http.Error(w, "start_date is required", http.StatusBadRequest)
			return
		}
		if req.SquareFootage <= 0 {
			http.Error(w, "square_footage must be positive", http.StatusBadRequest)
			return
		}
		if req.FoundationType == "" {
			http.Error(w, "foundation_type is required", http.StatusBadRequest)
			return
		}
		if req.Stories <= 0 {
			http.Error(w, "stories must be positive", http.StatusBadRequest)
			return
		}

		preview, err := h.previewSvc.GeneratePreview(req)
		if err != nil {
			slog.Error("preview generation failed", "error", err)
			http.Error(w, "preview generation failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(preview); err != nil {
			slog.Error("failed to encode preview response", "error", err)
		}
	}
}

// CompareScenarios handles POST /api/v1/schedule/compare
func (h *PreviewHandler) CompareScenarios() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxPreviewBodySize)

		var req service.ScenarioComparisonRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if len(req.Alternatives) == 0 {
			http.Error(w, "at least one alternative scenario is required", http.StatusBadRequest)
			return
		}
		if len(req.Alternatives) > 3 {
			http.Error(w, "maximum 3 alternative scenarios allowed", http.StatusBadRequest)
			return
		}

		result, err := h.previewSvc.CompareScenarios(req)
		if err != nil {
			slog.Error("scenario comparison failed", "error", err)
			http.Error(w, "scenario comparison failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			slog.Error("failed to encode comparison response", "error", err)
		}
	}
}
