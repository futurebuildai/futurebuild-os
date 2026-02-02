package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/httputil"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// getAuthOrgID extracts and validates the organization ID from JWT claims.
// Returns an error if claims are missing or orgID is invalid.
func getAuthOrgID(r *http.Request) (uuid.UUID, error) {
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		return uuid.Nil, errors.New("missing authentication context")
	}
	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		return uuid.Nil, errors.New("invalid OrgID in token")
	}
	return orgID, nil
}

type ProjectHandler struct {
	service service.ProjectServicer
}

func NewProjectHandler(s service.ProjectServicer) *ProjectHandler {
	return &ProjectHandler{service: s}
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	// Extract OrgID from authenticated JWT claims (secure, not from header)
	authOrgID, err := getAuthOrgID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var p models.Project
	// L7 Security: Prevent DoS via unbounded body
	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize) // 1MB limit
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
		if errors.Is(err, types.ErrConflict) {
			http.Error(w, "Project already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
	orgID, err := getAuthOrgID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	p, err := h.service.GetProject(r.Context(), projectID, orgID)
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(p)
}

func (h *ProjectHandler) GetProcurementItems(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// Extract OrgID from authenticated JWT claims
	orgID, err := getAuthOrgID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	items, err := h.service.ListProcurementItems(r.Context(), projectID, orgID)
	if err != nil {
		// Note: ListProcurementItems enforces multi-tenancy via JOIN;
		// it returns empty list (not ErrNotFound) if project doesn't exist or belongs to another org
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}
