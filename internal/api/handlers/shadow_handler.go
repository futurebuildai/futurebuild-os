package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/colton/futurebuild/internal/api/response"
	"github.com/colton/futurebuild/internal/futureshade/shadow"
	"github.com/colton/futurebuild/internal/middleware"
)

// ShadowHandler handles ShadowDocs endpoints.
// See SHADOW_VIEWER_specs.md Section 3.2
type ShadowHandler struct {
	service *shadow.DocsService
}

// NewShadowHandler creates a new ShadowHandler.
func NewShadowHandler(service *shadow.DocsService) *ShadowHandler {
	return &ShadowHandler{service: service}
}

// GetTree returns the docs/specs file tree.
// GET /api/v1/shadow/docs/tree
func (h *ShadowHandler) GetTree(w http.ResponseWriter, r *http.Request) {
	// Handle nil service (Fail Open)
	if h.service == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(shadow.TreeResponse{
			Roots: []shadow.TreeNode{},
		})
		return
	}

	tree, err := h.service.GetTree(r.Context())
	if err != nil {
		slog.Error("shadow: tree failed", "error", err)
		response.JSONError(w, http.StatusInternalServerError, "Failed to get file tree")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tree)
}

// GetContent returns file content with path validation.
// GET /api/v1/shadow/docs/content?path={path}
// SECURITY: Path traversal protection per SHADOW_VIEWER_specs.md Section 6.2
func (h *ShadowHandler) GetContent(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		response.JSONError(w, http.StatusBadRequest, "path parameter is required")
		return
	}

	// Audit log
	claims, _ := middleware.GetClaims(r.Context())
	slog.Info("shadow: doc requested",
		"user_id", claims.UserID,
		"path", path,
	)

	// Handle nil service (Fail Closed)
	if h.service == nil {
		response.JSONError(w, http.StatusServiceUnavailable, "Shadow service unavailable")
		return
	}

	content, err := h.service.GetContent(r.Context(), path)
	if err != nil {
		switch err {
		case shadow.ErrPathTraversal:
			slog.Warn("shadow: path traversal attempt blocked",
				"user_id", claims.UserID,
				"path", path,
			)
			// SECURITY: Don't reveal specific error to client
			response.JSONError(w, http.StatusBadRequest, "Invalid path")
		case shadow.ErrInvalidPath:
			response.JSONError(w, http.StatusBadRequest, "Invalid path")
		case shadow.ErrFileNotFound:
			response.JSONError(w, http.StatusNotFound, "File not found")
		default:
			slog.Error("shadow: content failed", "error", err, "path", path)
			response.JSONError(w, http.StatusInternalServerError, "Failed to read file")
		}
		return
	}

	slog.Info("shadow: doc viewed",
		"user_id", claims.UserID,
		"path", content.Path,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(content)
}
