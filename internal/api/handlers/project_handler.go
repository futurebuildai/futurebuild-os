package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	service service.ProjectServicer
}

func NewProjectHandler(s service.ProjectServicer) *ProjectHandler {
	return &ProjectHandler{service: s}
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	// Extract OrgID from authenticated JWT claims (secure, not from header)
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		http.Error(w, "Internal server error: missing authentication context", http.StatusInternalServerError)
		return
	}
	authOrgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		http.Error(w, "Internal server error: invalid OrgID in token", http.StatusInternalServerError)
		return
	}

	var p models.Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Multi-Tenancy Enforcement: If body OrgID is empty, assign from claims.
	// If body OrgID differs from claims, deny the request.
	if p.OrgID == uuid.Nil {
		p.OrgID = authOrgID
	} else if p.OrgID != authOrgID {
		http.Error(w, "cannot create project for another organization", http.StatusForbidden)
		return
	}

	if p.Name == "" {
		http.Error(w, "project name is required", http.StatusBadRequest)
		return
	}

	if err := h.service.CreateProject(r.Context(), &p); err != nil {
		if err.Error() == "project already exists" { // Simplified check or use custom error type
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		// Better: check for "already exists" in error message
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(p)
}

func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// Extract OrgID from authenticated JWT claims (secure, not from header)
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		http.Error(w, "Internal server error: missing authentication context", http.StatusInternalServerError)
		return
	}
	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		http.Error(w, "Internal server error: invalid OrgID in token", http.StatusInternalServerError)
		return
	}

	p, err := h.service.GetProject(r.Context(), projectID, orgID)
	if err != nil {
		http.Error(w, "Project not found or access denied", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(p)
}
