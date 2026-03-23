package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// MaterialHandler handles material CRUD operations.
type MaterialHandler struct {
	materialService *service.MaterialService
}

// NewMaterialHandler creates a new material handler.
func NewMaterialHandler(svc *service.MaterialService) *MaterialHandler {
	return &MaterialHandler{materialService: svc}
}

// ListMaterials returns all materials for a project.
// GET /api/v1/projects/{id}/materials
func (h *MaterialHandler) ListMaterials(w http.ResponseWriter, r *http.Request) {
	projectID, orgID, err := extractProjectAndOrgIDs(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	materials, err := h.materialService.ListMaterials(r.Context(), projectID, orgID)
	if err != nil {
		http.Error(w, "failed to list materials", http.StatusInternalServerError)
		return
	}

	if materials == nil {
		materials = []models.ProjectMaterial{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(materials)
}

// CreateMaterial adds a manually-entered material.
// POST /api/v1/projects/{id}/materials
func (h *MaterialHandler) CreateMaterial(w http.ResponseWriter, r *http.Request) {
	projectID, orgID, err := extractProjectAndOrgIDs(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req models.CreateMaterialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.WBSPhaseCode == "" {
		http.Error(w, "name and wbs_phase_code are required", http.StatusBadRequest)
		return
	}

	// Convert to MaterialEstimate for SaveMaterials
	estimate := models.MaterialEstimate{
		Name:           req.Name,
		Category:       req.Category,
		WBSPhaseCode:   req.WBSPhaseCode,
		Quantity:       req.Quantity,
		Unit:           req.Unit,
		UnitCostCents:  req.UnitCostCents,
		TotalCostCents: 0, // Calculated by service
		Confidence:     1.0,
		Source:         "user",
	}

	if err := h.materialService.SaveMaterials(r.Context(), projectID, []models.MaterialEstimate{estimate}); err != nil {
		http.Error(w, "failed to save material", http.StatusInternalServerError)
		return
	}

	// Return the saved materials list (includes the new one)
	materials, err := h.materialService.ListMaterials(r.Context(), projectID, orgID)
	if err != nil {
		http.Error(w, "failed to retrieve materials", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(materials)
}

// UpdateMaterial updates a material (user edits to AI-extracted data).
// PUT /api/v1/projects/{id}/materials/{materialId}
func (h *MaterialHandler) UpdateMaterial(w http.ResponseWriter, r *http.Request) {
	_, orgID, err := extractProjectAndOrgIDs(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	materialIDStr := chi.URLParam(r, "materialId")
	materialID, err := uuid.Parse(materialIDStr)
	if err != nil {
		http.Error(w, "invalid material ID", http.StatusBadRequest)
		return
	}

	var updates models.MaterialUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	material, err := h.materialService.UpdateMaterial(r.Context(), materialID, orgID, updates)
	if errors.Is(err, types.ErrNotFound) {
		http.Error(w, "material not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to update material", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(material)
}

// DeleteMaterial removes a material.
// DELETE /api/v1/projects/{id}/materials/{materialId}
func (h *MaterialHandler) DeleteMaterial(w http.ResponseWriter, r *http.Request) {
	_, orgID, err := extractProjectAndOrgIDs(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	materialIDStr := chi.URLParam(r, "materialId")
	materialID, err := uuid.Parse(materialIDStr)
	if err != nil {
		http.Error(w, "invalid material ID", http.StatusBadRequest)
		return
	}

	err = h.materialService.DeleteMaterial(r.Context(), materialID, orgID)
	if errors.Is(err, types.ErrNotFound) {
		http.Error(w, "material not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to delete material", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
