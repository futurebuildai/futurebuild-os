package handlers

import (
	"net/http"

	"github.com/colton/futurebuild/internal/service"
)

// PortalDashboardHandler handles portal dashboard API for authenticated contacts
type PortalDashboardHandler struct {
	portalService *service.PortalService
}

// NewPortalDashboardHandler creates a new PortalDashboardHandler
func NewPortalDashboardHandler(portalService *service.PortalService) *PortalDashboardHandler {
	return &PortalDashboardHandler{portalService: portalService}
}

// ListProjects returns projects for the authenticated contact
func (h *PortalDashboardHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

// ListProjectTasks returns tasks for a specific project
func (h *PortalDashboardHandler) ListProjectTasks(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

// GetDependencies returns task dependencies for a project
func (h *PortalDashboardHandler) GetDependencies(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

// ListMessages returns messages for a project
func (h *PortalDashboardHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

// SendMessage sends a message to the project
func (h *PortalDashboardHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

// ListDocuments returns documents for a project
func (h *PortalDashboardHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

// UploadDocument uploads a document to the project
func (h *PortalDashboardHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

// ListInvoices returns invoices for a project
func (h *PortalDashboardHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

// UploadInvoice uploads an invoice to the project
func (h *PortalDashboardHandler) UploadInvoice(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}
