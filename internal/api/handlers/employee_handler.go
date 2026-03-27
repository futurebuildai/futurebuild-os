package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/colton/futurebuild/internal/api/response"
	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// EmployeeHandler handles employee management endpoints.
// See BACKEND_SCOPE.md Section 20.2
type EmployeeHandler struct {
	svc service.EmployeeServicer
}

// NewEmployeeHandler creates an EmployeeHandler.
func NewEmployeeHandler(svc service.EmployeeServicer) *EmployeeHandler {
	return &EmployeeHandler{svc: svc}
}

// ListEmployees handles GET /api/v1/employees.
func (h *EmployeeHandler) ListEmployees(w http.ResponseWriter, r *http.Request) {
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
	employees, err := h.svc.ListEmployees(r.Context(), orgID, status)
	if err != nil {
		slog.Error("employees: failed to list", "error", err, "org_id", orgID)
		response.JSONError(w, http.StatusInternalServerError, "failed to list employees")
		return
	}
	if employees == nil {
		employees = []models.Employee{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": employees})
}

// CreateEmployee handles POST /api/v1/employees.
func (h *EmployeeHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
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
	var emp models.Employee
	if err := json.NewDecoder(r.Body).Decode(&emp); err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if emp.FirstName == "" || emp.LastName == "" {
		response.JSONError(w, http.StatusBadRequest, "first_name and last_name are required")
		return
	}

	if err := h.svc.CreateEmployee(r.Context(), orgID, &emp); err != nil {
		slog.Error("employees: failed to create", "error", err, "org_id", orgID)
		response.JSONError(w, http.StatusInternalServerError, "failed to create employee")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": emp})
}

// GetEmployee handles GET /api/v1/employees/{id}.
func (h *EmployeeHandler) GetEmployee(w http.ResponseWriter, r *http.Request) {
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

	employeeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid employee ID")
		return
	}

	emp, err := h.svc.GetEmployee(r.Context(), employeeID, orgID)
	if err != nil {
		slog.Error("employees: failed to get", "error", err, "org_id", orgID, "employee_id", employeeID)
		response.JSONError(w, http.StatusNotFound, "employee not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": emp})
}

// UpdateEmployee handles PUT /api/v1/employees/{id}.
func (h *EmployeeHandler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
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

	employeeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid employee ID")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var emp models.Employee
	if err := json.NewDecoder(r.Body).Decode(&emp); err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updated, err := h.svc.UpdateEmployee(r.Context(), employeeID, orgID, &emp)
	if err != nil {
		slog.Error("employees: failed to update", "error", err, "org_id", orgID, "employee_id", employeeID)
		response.JSONError(w, http.StatusInternalServerError, "failed to update employee")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": updated})
}

// LogTime handles POST /api/v1/employees/{id}/time-logs.
func (h *EmployeeHandler) LogTime(w http.ResponseWriter, r *http.Request) {
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

	employeeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid employee ID")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var log models.TimeLog
	if err := json.NewDecoder(r.Body).Decode(&log); err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	log.EmployeeID = employeeID
	if err := h.svc.LogTime(r.Context(), &log); err != nil {
		slog.Error("employees: failed to log time", "error", err, "employee_id", employeeID)
		response.JSONError(w, http.StatusInternalServerError, "failed to log time")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": log})
}

// GetTimeLogs handles GET /api/v1/employees/{id}/time-logs.
func (h *EmployeeHandler) GetTimeLogs(w http.ResponseWriter, r *http.Request) {
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

	employeeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid employee ID")
		return
	}

	logs, err := h.svc.GetTimeLogs(r.Context(), employeeID)
	if err != nil {
		slog.Error("employees: failed to get time logs", "error", err, "employee_id", employeeID)
		response.JSONError(w, http.StatusInternalServerError, "failed to get time logs")
		return
	}
	if logs == nil {
		logs = []models.TimeLog{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": logs})
}

// AddCertification handles POST /api/v1/employees/{id}/certifications.
func (h *EmployeeHandler) AddCertification(w http.ResponseWriter, r *http.Request) {
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

	employeeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid employee ID")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var cert models.Certification
	if err := json.NewDecoder(r.Body).Decode(&cert); err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cert.EmployeeID = employeeID
	if err := h.svc.AddCertification(r.Context(), &cert); err != nil {
		slog.Error("employees: failed to add certification", "error", err, "employee_id", employeeID)
		response.JSONError(w, http.StatusInternalServerError, "failed to add certification")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": cert})
}

// ListCertifications handles GET /api/v1/employees/{id}/certifications.
func (h *EmployeeHandler) ListCertifications(w http.ResponseWriter, r *http.Request) {
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

	employeeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid employee ID")
		return
	}

	certs, err := h.svc.ListCertifications(r.Context(), employeeID)
	if err != nil {
		slog.Error("employees: failed to list certifications", "error", err, "employee_id", employeeID)
		response.JSONError(w, http.StatusInternalServerError, "failed to list certifications")
		return
	}
	if certs == nil {
		certs = []models.Certification{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": certs})
}

// GetExpiringCertifications handles GET /api/v1/certifications/expiring?within_days=.
func (h *EmployeeHandler) GetExpiringCertifications(w http.ResponseWriter, r *http.Request) {
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

	withinDays := 30
	if d := r.URL.Query().Get("within_days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 365 {
			withinDays = parsed
		}
	}

	certs, err := h.svc.GetExpiringCertifications(r.Context(), orgID, withinDays)
	if err != nil {
		slog.Error("employees: failed to get expiring certs", "error", err, "org_id", orgID)
		response.JSONError(w, http.StatusInternalServerError, "failed to get expiring certifications")
		return
	}
	if certs == nil {
		certs = []models.Certification{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": certs})
}

// ApproveTimeLog handles POST /api/v1/time-logs/{id}/approve.
func (h *EmployeeHandler) ApproveTimeLog(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		response.JSONError(w, http.StatusUnauthorized, "missing auth context")
		return
	}

	approverID, err := uuid.Parse(claims.UserID)
	if err != nil {
		response.JSONError(w, http.StatusUnauthorized, "invalid user context")
		return
	}

	logID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid time log ID")
		return
	}

	if err := h.svc.ApproveTimeLog(r.Context(), logID, approverID); err != nil {
		slog.Error("employees: failed to approve time log", "error", err, "log_id", logID)
		response.JSONError(w, http.StatusInternalServerError, "failed to approve time log")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "approved"})
}

// GetLaborBurden handles GET /api/v1/projects/{id}/labor-burden.
func (h *EmployeeHandler) GetLaborBurden(w http.ResponseWriter, r *http.Request) {
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

	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	totalCents, err := h.svc.CalculateLaborBurden(r.Context(), projectID)
	if err != nil {
		slog.Error("employees: failed to calculate labor burden", "error", err, "project_id", projectID)
		response.JSONError(w, http.StatusInternalServerError, "failed to calculate labor burden")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": map[string]int64{"total_labor_cost_cents": totalCents},
	})
}
