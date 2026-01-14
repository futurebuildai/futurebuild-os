package handlers

import (
	"encoding/json"
	"net/http"

	"strings"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	service *service.ProjectService
}

func NewProjectHandler(s *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: s}
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.Header.Get("X-Org-ID")
	if orgIDStr == "" {
		http.Error(w, "X-Org-ID header is required for multi-tenancy validation", http.StatusBadRequest)
		return
	}
	authOrgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "Invalid X-Org-ID", http.StatusBadRequest)
		return
	}

	var p models.Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Rigid Multi-Tenancy Enforcement: Header MUST match body
	if p.OrgID != uuid.Nil && p.OrgID != authOrgID {
		http.Error(w, "Authorized OrgID does not match request body", http.StatusForbidden)
		return
	}
	p.OrgID = authOrgID

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

	// Placeholder for multi-tenancy until Auth is implemented
	orgIDStr := r.Header.Get("X-Org-ID")
	if orgIDStr == "" {
		http.Error(w, "X-Org-ID header is required for multi-tenancy validation", http.StatusBadRequest)
		return
	}
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "Invalid X-Org-ID", http.StatusBadRequest)
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
