package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// VisionHandler handles vision verification endpoints.
// See PRODUCTION_PLAN.md Step 40
type VisionHandler struct {
	visionService *service.VisionService
	db            *pgxpool.Pool
}

// NewVisionHandler creates a new VisionHandler.
func NewVisionHandler(vs *service.VisionService, db *pgxpool.Pool) *VisionHandler {
	return &VisionHandler{
		visionService: vs,
		db:            db,
	}
}

// VerifyTaskRequest represents the request body for POST /tasks/{task_id}/verify.
// See PRODUCTION_PLAN.md Step 40
type VerifyTaskRequest struct {
	ImageURL        string `json:"image_url"`
	TaskDescription string `json:"task_description"`
}

// VerifyTaskResponse represents the response for POST /tasks/{task_id}/verify.
type VerifyTaskResponse struct {
	IsVerified bool    `json:"is_verified"`
	Confidence float64 `json:"confidence"`
}

// VerifyTask handles POST /api/v1/projects/{id}/tasks/{task_id}/verify.
// Completes the "Site Photo Verification Flow" from Step 40.
// See PRODUCTION_PLAN.md Step 40, DATA_SPINE_SPEC.md Section 3.3
func (h *VisionHandler) VerifyTask(w http.ResponseWriter, r *http.Request) {
	// 1. Extract IDs and enforce multi-tenancy
	projectID, orgID, err := extractProjectAndOrgIDs(r)
	if err != nil {
		slog.Warn("vision: invalid project/org IDs", "error", err, "method", r.Method)
		http.Error(w, "Invalid project or organization ID", http.StatusBadRequest)
		return
	}

	taskIDStr := chi.URLParam(r, "task_id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		slog.Warn("vision: invalid task ID", "raw_task_id", taskIDStr, "error", err)
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// 2. Parse request body (L7: limit body size)
	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var req VerifyTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("vision: invalid request body", "error", err, "task_id", taskID)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ImageURL == "" {
		slog.Warn("vision: image_url missing", "task_id", taskID)
		http.Error(w, "image_url is required", http.StatusBadRequest)
		return
	}

	if req.TaskDescription == "" {
		slog.Warn("vision: task_description missing", "task_id", taskID)
		http.Error(w, "task_description is required", http.StatusBadRequest)
		return
	}

	slog.Info("vision: verifying task",
		"task_id", taskID, "project_id", projectID, "org_id", orgID)

	// 3. Call VisionService with Persistence
	// This satisfies CTO Audit requirement: "Database is State"
	isVerified, confidence, err := h.visionService.VerifyAndPersistTask(
		r.Context(),
		h.db,
		taskID,
		projectID,
		orgID,
		req.ImageURL,
		req.TaskDescription,
	)
	if err != nil {
		slog.Error("vision: verification failed",
			"task_id", taskID, "project_id", projectID, "error", err)
		http.Error(w, "Vision verification failed", http.StatusInternalServerError)
		return
	}

	slog.Info("vision: verification completed",
		"task_id", taskID, "is_verified", isVerified, "confidence", confidence)

	// 4. Return result
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(VerifyTaskResponse{
		IsVerified: isVerified,
		Confidence: confidence,
	})
}

// ExtractDocument handles POST /api/v1/vision/extract.
// Sprint 2.1 Task 2.1.1: Multipart upload endpoint for construction plan extraction.
// Accepts PDF/image files, calls VisionService.ParseDocument, returns VisionExtractionResponse.
func (h *VisionHandler) ExtractDocument(w http.ResponseWriter, r *http.Request) {
	// 1. Limit request size (20MB max for construction plans)
	const maxFileSize = 20 * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize+1024) // +1KB for form fields

	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		slog.Warn("vision: extract multipart parse error", "error", err)
		http.Error(w, "File too large or invalid upload (max 20MB)", http.StatusBadRequest)
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()

	// 2. Extract uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		slog.Warn("vision: extract file missing", "error", err)
		http.Error(w, "file field is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 3. Validate MIME type
	mimeType := header.Header.Get("Content-Type")
	if !isValidExtractionMIME(mimeType) {
		slog.Warn("vision: extract invalid mime type", "mime_type", mimeType)
		http.Error(w, "Invalid file type. Accepted: PDF, PNG, JPG, WebP", http.StatusBadRequest)
		return
	}

	// 4. Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		slog.Error("vision: extract read file error", "error", err)
		http.Error(w, "Failed to read uploaded file", http.StatusInternalServerError)
		return
	}

	if len(fileData) == 0 {
		http.Error(w, "Uploaded file is empty", http.StatusBadRequest)
		return
	}

	slog.Info("vision: extracting from uploaded document",
		"file_name", header.Filename,
		"file_size", len(fileData),
		"mime_type", mimeType,
	)

	// 5. Call VisionService.ParseDocument
	resp, err := h.visionService.ParseDocument(r.Context(), fileData, mimeType)
	if err != nil {
		slog.Error("vision: extraction failed",
			"file_name", header.Filename,
			"error", err,
		)
		http.Error(w, "Document extraction failed. Please try a clearer scan.", http.StatusInternalServerError)
		return
	}

	slog.Info("vision: extraction success",
		"file_name", header.Filename,
		"extracted_fields", len(resp.ExtractedValues),
		"overall_confidence", resp.ConfidenceReport.OverallConfidence,
	)

	// 6. Return VisionExtractionResponse
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("vision: failed to encode response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// isValidExtractionMIME validates MIME types for construction plan uploads.
func isValidExtractionMIME(mimeType string) bool {
	// Strip parameters: "image/jpeg; charset=utf-8" → "image/jpeg"
	normalized := strings.TrimSpace(strings.SplitN(mimeType, ";", 2)[0])
	allowed := map[string]bool{
		"image/jpeg":      true,
		"image/jpg":       true,
		"image/png":       true,
		"image/webp":      true,
		"application/pdf": true,
	}
	return allowed[normalized]
}
