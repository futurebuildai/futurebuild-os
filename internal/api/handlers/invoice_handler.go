package handlers

import (
	"net/http"

	"github.com/colton/futurebuild/internal/service"
)

// InvoiceHandler handles interactive invoice operations
type InvoiceHandler struct {
	invoiceService *service.InvoiceService
}

// NewInvoiceHandler creates a new InvoiceHandler
func NewInvoiceHandler(invoiceService *service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{invoiceService: invoiceService}
}

// GetInvoice returns invoice details
func (h *InvoiceHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

// UpdateInvoice updates an invoice
func (h *InvoiceHandler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

// ApproveInvoice approves an invoice
func (h *InvoiceHandler) ApproveInvoice(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

// RejectInvoice rejects an invoice
func (h *InvoiceHandler) RejectInvoice(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}
