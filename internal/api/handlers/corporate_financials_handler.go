package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/colton/futurebuild/internal/api/response"
	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/service"
	"github.com/google/uuid"
)

// CorporateFinancialsHandler handles corporate financial endpoints.
// See BACKEND_SCOPE.md Section 20.1
type CorporateFinancialsHandler struct {
	svc service.CorporateFinancialsServicer
}

// NewCorporateFinancialsHandler creates a CorporateFinancialsHandler.
func NewCorporateFinancialsHandler(svc service.CorporateFinancialsServicer) *CorporateFinancialsHandler {
	return &CorporateFinancialsHandler{svc: svc}
}

// GetCorporateBudget handles GET /api/v1/corporate/budgets?year=&quarter=.
func (h *CorporateFinancialsHandler) GetCorporateBudget(w http.ResponseWriter, r *http.Request) {
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

	yearStr := r.URL.Query().Get("year")
	quarterStr := r.URL.Query().Get("quarter")
	if yearStr == "" || quarterStr == "" {
		response.JSONError(w, http.StatusBadRequest, "year and quarter are required")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid year")
		return
	}
	quarter, err := strconv.Atoi(quarterStr)
	if err != nil || quarter < 1 || quarter > 4 {
		response.JSONError(w, http.StatusBadRequest, "quarter must be 1-4")
		return
	}

	budget, err := h.svc.GetCorporateBudget(r.Context(), orgID, year, quarter)
	if err != nil {
		slog.Error("corporate: failed to get budget", "error", err, "org_id", orgID)
		response.JSONError(w, http.StatusInternalServerError, "failed to get corporate budget")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": budget})
}

// RollupBudget handles POST /api/v1/corporate/budgets/rollup.
func (h *CorporateFinancialsHandler) RollupBudget(w http.ResponseWriter, r *http.Request) {
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

	type rollupRequest struct {
		Year    int `json:"year"`
		Quarter int `json:"quarter"`
	}

	var req rollupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Year < 2000 || req.Year > 2100 {
		response.JSONError(w, http.StatusBadRequest, "invalid year")
		return
	}
	if req.Quarter < 1 || req.Quarter > 4 {
		response.JSONError(w, http.StatusBadRequest, "quarter must be 1-4")
		return
	}

	budget, err := h.svc.RollupCorporateBudget(r.Context(), orgID, req.Year, req.Quarter)
	if err != nil {
		slog.Error("corporate: rollup failed", "error", err, "org_id", orgID)
		response.JSONError(w, http.StatusInternalServerError, "failed to rollup budget")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": budget})
}

// GetARAging handles GET /api/v1/corporate/ar-aging.
func (h *CorporateFinancialsHandler) GetARAging(w http.ResponseWriter, r *http.Request) {
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

	snapshot, err := h.svc.CalculateARAging(r.Context(), orgID)
	if err != nil {
		slog.Error("corporate: AR aging failed", "error", err, "org_id", orgID)
		response.JSONError(w, http.StatusInternalServerError, "failed to calculate AR aging")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": snapshot})
}

// ListGLSyncLogs handles GET /api/v1/corporate/gl-sync.
func (h *CorporateFinancialsHandler) ListGLSyncLogs(w http.ResponseWriter, r *http.Request) {
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

	logs, err := h.svc.ListGLSyncLogs(r.Context(), orgID)
	if err != nil {
		slog.Error("corporate: failed to list GL sync logs", "error", err, "org_id", orgID)
		response.JSONError(w, http.StatusInternalServerError, "failed to list GL sync logs")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": logs})
}

// CreateGLSyncLog handles POST /api/v1/corporate/gl-sync.
func (h *CorporateFinancialsHandler) CreateGLSyncLog(w http.ResponseWriter, r *http.Request) {
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

	type syncRequest struct {
		SyncType string `json:"sync_type"`
	}

	var req syncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SyncType == "" {
		response.JSONError(w, http.StatusBadRequest, "sync_type is required")
		return
	}

	log, err := h.svc.CreateGLSyncLog(r.Context(), orgID, req.SyncType)
	if err != nil {
		slog.Error("corporate: failed to create GL sync log", "error", err, "org_id", orgID)
		response.JSONError(w, http.StatusInternalServerError, "failed to create GL sync log")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": log})
}
