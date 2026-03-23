package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"

	"github.com/colton/futurebuild/internal/api/response"
	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// progressFeedWriter is satisfied by *service.FeedService.
type progressFeedWriter interface {
	WriteCard(ctx context.Context, card *models.FeedCard) error
}

// ProgressPhotoHandler handles progress photo upload and classification.
type ProgressPhotoHandler struct {
	visionSvc  *service.VisionService
	feedWriter progressFeedWriter
}

// NewProgressPhotoHandler creates a new progress photo handler.
func NewProgressPhotoHandler(
	visionSvc *service.VisionService,
	feedWriter progressFeedWriter,
) *ProgressPhotoHandler {
	return &ProgressPhotoHandler{
		visionSvc:  visionSvc,
		feedWriter: feedWriter,
	}
}

// HandleUpload processes a photo upload, classifies it, and creates an approval card.
// POST /api/v1/projects/{projectId}/progress/photo
func (h *ProgressPhotoHandler) HandleUpload() http.HandlerFunc {
	const maxUploadSize = 10 << 20 // 10MB

	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			response.JSONError(w, http.StatusBadRequest, "invalid project_id")
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

		// Parse multipart form
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			response.JSONError(w, http.StatusBadRequest, "file too large or invalid form")
			return
		}

		file, header, err := r.FormFile("photo")
		if err != nil {
			response.JSONError(w, http.StatusBadRequest, "missing photo field")
			return
		}
		defer file.Close()

		// Read file bytes
		imageBytes, err := io.ReadAll(file)
		if err != nil {
			response.JSONError(w, http.StatusInternalServerError, "failed to read file")
			return
		}

		mimeType := header.Header.Get("Content-Type")
		if mimeType == "" {
			mimeType = "image/jpeg"
		}

		// Validate MIME type against allowed image types
		allowedTypes := []string{"image/jpeg", "image/png", "image/webp", "image/heic"}
		if !slices.Contains(allowedTypes, mimeType) {
			response.JSONError(w, http.StatusBadRequest, "unsupported image type; accepted: jpeg, png, webp, heic")
			return
		}

		// Classify the photo
		classification, err := h.visionSvc.ClassifyProgressPhoto(r.Context(), imageBytes, mimeType)
		if err != nil {
			slog.Error("progress photo classification failed", "error", err, "project_id", projectID)
			response.JSONError(w, http.StatusInternalServerError, "photo classification failed")
			return
		}

		// Create approval card for updating task progress
		if h.feedWriter != nil && classification.Confidence >= 0.6 {
			headline := fmt.Sprintf("Photo shows %s at ~%d%% complete",
				classification.DetectedPhase, classification.EstimatedPercent)

			body := fmt.Sprintf("AI analyzed a site photo and detected %s phase (WBS %s) at approximately %d%% completion.",
				classification.DetectedPhase, classification.WBSCode, classification.EstimatedPercent)
			if len(classification.VisibleElements) > 0 {
				body += "\n\nVisible elements:"
				for _, elem := range classification.VisibleElements {
					body += fmt.Sprintf("\n- %s", elem)
				}
			}

			agentSource := "ProgressPhotoVerification"
			consequence := fmt.Sprintf("Update task %s progress to %d%%", classification.WBSCode, classification.EstimatedPercent)

			card := &models.FeedCard{
				ID:          uuid.New(),
				OrgID:       orgID,
				ProjectID:   projectID,
				CardType:    models.FeedCardAgentApproval,
				Priority:    models.FeedCardPriorityNormal,
				Headline:    headline,
				Body:        body,
				Consequence: &consequence,
				Horizon:     models.FeedCardHorizonToday,
				AgentSource: &agentSource,
				Actions: []models.FeedCardAction{
					{ID: "approve_agent_action", Label: "Update Progress", Style: "primary"},
					{ID: "reject_agent_action", Label: "Ignore", Style: "secondary"},
				},
			}

			if writeErr := h.feedWriter.WriteCard(r.Context(), card); writeErr != nil {
				slog.Warn("failed to write progress photo card", "error", writeErr)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"classification": classification,
			},
		})
	}
}
