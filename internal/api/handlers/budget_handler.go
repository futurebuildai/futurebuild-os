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

// BudgetHandler handles budget CRUD and financial summary operations.
type BudgetHandler struct {
	budgetService   *service.BudgetService
	materialService *service.MaterialService
}

// NewBudgetHandler creates a new budget handler.
func NewBudgetHandler(bs *service.BudgetService, ms *service.MaterialService) *BudgetHandler {
	return &BudgetHandler{budgetService: bs, materialService: ms}
}

// GetBudgetBreakdown returns per-phase budget data.
// GET /api/v1/projects/{id}/budget
func (h *BudgetHandler) GetBudgetBreakdown(w http.ResponseWriter, r *http.Request) {
	projectID, orgID, err := extractProjectAndOrgIDs(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	budgets, err := h.budgetService.GetBudgetBreakdown(r.Context(), projectID, orgID)
	if err != nil {
		http.Error(w, "failed to get budget breakdown", http.StatusInternalServerError)
		return
	}

	if budgets == nil {
		budgets = []models.ProjectBudget{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(budgets)
}

// UpdateBudgetPhase allows user to override a phase budget estimate.
// PUT /api/v1/projects/{id}/budget/{budgetId}
func (h *BudgetHandler) UpdateBudgetPhase(w http.ResponseWriter, r *http.Request) {
	_, orgID, err := extractProjectAndOrgIDs(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	budgetIDStr := chi.URLParam(r, "budgetId")
	budgetID, err := uuid.Parse(budgetIDStr)
	if err != nil {
		http.Error(w, "invalid budget ID", http.StatusBadRequest)
		return
	}

	var req struct {
		EstimatedAmountCents int64 `json:"estimated_amount_cents"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	budget, err := h.budgetService.UpdateBudgetPhase(r.Context(), budgetID, orgID, req.EstimatedAmountCents)
	if errors.Is(err, types.ErrNotFound) {
		http.Error(w, "budget not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to update budget", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(budget)
}

// SeedBudget creates initial budget from material estimates.
// POST /api/v1/projects/{id}/budget/seed
func (h *BudgetHandler) SeedBudget(w http.ResponseWriter, r *http.Request) {
	projectID, _, err := extractProjectAndOrgIDs(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req models.BudgetSeedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Extract project attributes for cost index calculations
	// Use defaults if not provided in the seed request
	gsf := 2250.0
	foundationType := "slab"
	stories := 1
	multiplier := req.RegionalMultiplier
	if multiplier <= 0 {
		multiplier = 1.0
	}

	estimate, err := h.budgetService.SeedBudget(
		r.Context(), projectID, req.Materials,
		gsf, foundationType, stories, multiplier,
	)
	if err != nil {
		http.Error(w, "failed to seed budget", http.StatusInternalServerError)
		return
	}

	// Also persist the materials
	if len(req.Materials) > 0 {
		if err := h.materialService.SaveMaterials(r.Context(), projectID, req.Materials); err != nil {
			// Non-fatal: budget was seeded, materials failed to save
			// Log but continue
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(estimate)
}

// GetFinancialSummary returns budget vs spend for a project.
// GET /api/v1/projects/{id}/financials/summary
func (h *BudgetHandler) GetFinancialSummary(w http.ResponseWriter, r *http.Request) {
	projectID, orgID, err := extractProjectAndOrgIDs(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	summary, err := h.budgetService.GetFinancialSummary(r.Context(), projectID, orgID)
	if errors.Is(err, types.ErrNotFound) {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to get financial summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summary)
}

// GetGlobalSummary returns aggregated financials across all projects.
// GET /api/v1/financials/summary
func (h *BudgetHandler) GetGlobalSummary(w http.ResponseWriter, r *http.Request) {
	orgID, err := getAuthOrgID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	summary, err := h.budgetService.GetGlobalFinancialSummary(r.Context(), orgID)
	if err != nil {
		http.Error(w, "failed to get global financial summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summary)
}
