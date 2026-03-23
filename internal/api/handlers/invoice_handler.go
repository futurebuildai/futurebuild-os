package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// InvoiceHandler handles interactive invoice operations.
// See PHASE_13_PRD.md Step 82: Interactive Invoice
type InvoiceHandler struct {
	invoiceService *service.InvoiceService
}

// NewInvoiceHandler creates a new InvoiceHandler.
func NewInvoiceHandler(invoiceService *service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{invoiceService: invoiceService}
}

// GetInvoice returns invoice details.
// GET /api/v1/invoices/{id}
func (h *InvoiceHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	orgID, err := getAuthOrgID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	invoiceIDStr := chi.URLParam(r, "id")
	invoiceID, err := uuid.Parse(invoiceIDStr)
	if err != nil {
		http.Error(w, "invalid invoice ID", http.StatusBadRequest)
		return
	}

	invoice, err := h.invoiceService.GetInvoice(r.Context(), invoiceID, orgID)
	if err != nil {
		http.Error(w, "invoice not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(invoice)
}

// UpdateInvoice updates an invoice's line items.
// PUT /api/v1/invoices/{id}
func (h *InvoiceHandler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	orgID, err := getAuthOrgID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	invoiceIDStr := chi.URLParam(r, "id")
	invoiceID, err := uuid.Parse(invoiceIDStr)
	if err != nil {
		http.Error(w, "invalid invoice ID", http.StatusBadRequest)
		return
	}

	var req struct {
		LineItems []models.LineItem `json:"line_items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	invoice, err := h.invoiceService.UpdateInvoiceItems(r.Context(), invoiceID, orgID, req.LineItems)
	if errors.Is(err, service.ErrInvoiceNotEditable) {
		http.Error(w, "invoice is not editable", http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "failed to update invoice", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(invoice)
}

// ApproveInvoice approves an invoice.
// POST /api/v1/invoices/{id}/approve
func (h *InvoiceHandler) ApproveInvoice(w http.ResponseWriter, r *http.Request) {
	orgID, err := getAuthOrgID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		http.Error(w, "missing authentication context", http.StatusUnauthorized)
		return
	}

	invoiceIDStr := chi.URLParam(r, "id")
	invoiceID, err := uuid.Parse(invoiceIDStr)
	if err != nil {
		http.Error(w, "invalid invoice ID", http.StatusBadRequest)
		return
	}

	invoice, err := h.invoiceService.ApproveInvoice(r.Context(), invoiceID, orgID, claims.UserID)
	if errors.Is(err, service.ErrInvoiceNotApprovable) {
		http.Error(w, "invoice cannot be approved", http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "failed to approve invoice", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(invoice)
}

// RejectInvoice rejects an invoice.
// POST /api/v1/invoices/{id}/reject
func (h *InvoiceHandler) RejectInvoice(w http.ResponseWriter, r *http.Request) {
	orgID, err := getAuthOrgID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		http.Error(w, "missing authentication context", http.StatusUnauthorized)
		return
	}

	invoiceIDStr := chi.URLParam(r, "id")
	invoiceID, err := uuid.Parse(invoiceIDStr)
	if err != nil {
		http.Error(w, "invalid invoice ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	invoice, err := h.invoiceService.RejectInvoice(r.Context(), invoiceID, orgID, claims.UserID, req.Reason)
	if errors.Is(err, service.ErrInvoiceNotRejectable) {
		http.Error(w, "invoice cannot be rejected", http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "failed to reject invoice", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(invoice)
}
