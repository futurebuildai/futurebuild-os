package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/colton/futurebuild/internal/api/response"
	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// FleetHandler handles fleet and equipment management endpoints.
// See BACKEND_SCOPE.md Section 20.3
type FleetHandler struct {
	svc service.FleetServicer
}

// NewFleetHandler creates a FleetHandler.
func NewFleetHandler(svc service.FleetServicer) *FleetHandler {
	return &FleetHandler{svc: svc}
}

// ListFleetAssets handles GET /api/v1/fleet.
func (h *FleetHandler) ListFleetAssets(w http.ResponseWriter, r *http.Request) {
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

	status := r.URL.Query().Get("status")
	assetType := r.URL.Query().Get("asset_type")

	assets, err := h.svc.ListFleetAssets(r.Context(), orgID, status, assetType)
	if err != nil {
		slog.Error("fleet: failed to list assets", "error", err, "org_id", orgID)
		response.JSONError(w, http.StatusInternalServerError, "failed to list fleet assets")
		return
	}
	if assets == nil {
		assets = []models.FleetAsset{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": assets})
}

// CreateFleetAsset handles POST /api/v1/fleet.
func (h *FleetHandler) CreateFleetAsset(w http.ResponseWriter, r *http.Request) {
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

	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var asset models.FleetAsset
	if err := json.NewDecoder(r.Body).Decode(&asset); err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if asset.AssetNumber == "" || asset.AssetType == "" {
		response.JSONError(w, http.StatusBadRequest, "asset_number and asset_type are required")
		return
	}

	if err := h.svc.CreateFleetAsset(r.Context(), orgID, &asset); err != nil {
		slog.Error("fleet: failed to create asset", "error", err, "org_id", orgID)
		response.JSONError(w, http.StatusInternalServerError, "failed to create fleet asset")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": asset})
}

// GetFleetAsset handles GET /api/v1/fleet/{id}.
func (h *FleetHandler) GetFleetAsset(w http.ResponseWriter, r *http.Request) {
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

	assetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid asset ID")
		return
	}

	asset, err := h.svc.GetFleetAsset(r.Context(), assetID, orgID)
	if err != nil {
		slog.Error("fleet: failed to get asset", "error", err, "org_id", orgID, "asset_id", assetID)
		response.JSONError(w, http.StatusNotFound, "fleet asset not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": asset})
}

// UpdateFleetAsset handles PUT /api/v1/fleet/{id}.
func (h *FleetHandler) UpdateFleetAsset(w http.ResponseWriter, r *http.Request) {
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

	assetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid asset ID")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var asset models.FleetAsset
	if err := json.NewDecoder(r.Body).Decode(&asset); err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updated, err := h.svc.UpdateFleetAsset(r.Context(), assetID, orgID, &asset)
	if err != nil {
		slog.Error("fleet: failed to update asset", "error", err, "org_id", orgID, "asset_id", assetID)
		response.JSONError(w, http.StatusInternalServerError, "failed to update fleet asset")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": updated})
}

// AllocateEquipment handles POST /api/v1/fleet/{id}/allocate.
func (h *FleetHandler) AllocateEquipment(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		response.JSONError(w, http.StatusUnauthorized, "missing auth context")
		return
	}

	_, err = uuid.Parse(claims.OrgID)
	if err != nil {
		response.JSONError(w, http.StatusUnauthorized, "invalid org context")
		return
	}

	assetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid asset ID")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var alloc models.EquipmentAllocation
	if err := json.NewDecoder(r.Body).Decode(&alloc); err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	alloc.AssetID = assetID
	if alloc.ProjectID == uuid.Nil {
		response.JSONError(w, http.StatusBadRequest, "project_id is required")
		return
	}

	if err := h.svc.AllocateEquipment(r.Context(), &alloc); err != nil {
		slog.Error("fleet: allocation failed", "error", err, "asset_id", assetID)
		response.JSONError(w, http.StatusConflict, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": alloc})
}

// CheckAvailability handles GET /api/v1/fleet/{id}/availability?from=&to=.
func (h *FleetHandler) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		response.JSONError(w, http.StatusUnauthorized, "missing auth context")
		return
	}

	_, err = uuid.Parse(claims.OrgID)
	if err != nil {
		response.JSONError(w, http.StatusUnauthorized, "invalid org context")
		return
	}

	assetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid asset ID")
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	if fromStr == "" || toStr == "" {
		response.JSONError(w, http.StatusBadRequest, "from and to dates are required (YYYY-MM-DD)")
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid from date format (use YYYY-MM-DD)")
		return
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid to date format (use YYYY-MM-DD)")
		return
	}

	available, err := h.svc.CheckEquipmentAvailability(r.Context(), assetID, from, to)
	if err != nil {
		slog.Error("fleet: availability check failed", "error", err, "asset_id", assetID)
		response.JSONError(w, http.StatusInternalServerError, "failed to check availability")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": map[string]bool{"available": available},
	})
}

// LogMaintenance handles POST /api/v1/fleet/{id}/maintenance.
func (h *FleetHandler) LogMaintenance(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		response.JSONError(w, http.StatusUnauthorized, "missing auth context")
		return
	}

	_, err = uuid.Parse(claims.OrgID)
	if err != nil {
		response.JSONError(w, http.StatusUnauthorized, "invalid org context")
		return
	}

	assetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid asset ID")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var log models.MaintenanceLog
	if err := json.NewDecoder(r.Body).Decode(&log); err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	log.AssetID = assetID
	if err := h.svc.LogMaintenance(r.Context(), &log); err != nil {
		slog.Error("fleet: failed to log maintenance", "error", err, "asset_id", assetID)
		response.JSONError(w, http.StatusInternalServerError, "failed to log maintenance")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": log})
}

// GetMaintenanceHistory handles GET /api/v1/fleet/{id}/maintenance.
func (h *FleetHandler) GetMaintenanceHistory(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		response.JSONError(w, http.StatusUnauthorized, "missing auth context")
		return
	}

	_, err = uuid.Parse(claims.OrgID)
	if err != nil {
		response.JSONError(w, http.StatusUnauthorized, "invalid org context")
		return
	}

	assetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid asset ID")
		return
	}

	logs, err := h.svc.GetMaintenanceHistory(r.Context(), assetID)
	if err != nil {
		slog.Error("fleet: failed to get maintenance history", "error", err, "asset_id", assetID)
		response.JSONError(w, http.StatusInternalServerError, "failed to get maintenance history")
		return
	}
	if logs == nil {
		logs = []models.MaintenanceLog{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": logs})
}

// GetProjectEquipment handles GET /api/v1/projects/{id}/equipment.
func (h *FleetHandler) GetProjectEquipment(w http.ResponseWriter, r *http.Request) {
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

	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	allocs, err := h.svc.GetProjectEquipment(r.Context(), projectID, orgID)
	if err != nil {
		slog.Error("fleet: failed to get project equipment", "error", err, "project_id", projectID)
		response.JSONError(w, http.StatusInternalServerError, "failed to get project equipment")
		return
	}
	if allocs == nil {
		allocs = []models.EquipmentAllocation{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": allocs})
}

// GetUpcomingMaintenance handles GET /api/v1/maintenance/upcoming?within_days=.
func (h *FleetHandler) GetUpcomingMaintenance(w http.ResponseWriter, r *http.Request) {
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

	withinDays := 14
	if d := r.URL.Query().Get("within_days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 365 {
			withinDays = parsed
		}
	}

	logs, err := h.svc.GetUpcomingMaintenance(r.Context(), orgID, withinDays)
	if err != nil {
		slog.Error("fleet: failed to get upcoming maintenance", "error", err, "org_id", orgID)
		response.JSONError(w, http.StatusInternalServerError, "failed to get upcoming maintenance")
		return
	}
	if logs == nil {
		logs = []models.MaintenanceLog{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": logs})
}
