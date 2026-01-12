package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/colton/futurebuild/internal/service"
	"github.com/google/uuid"
)

// DocumentHandler handles requests related to document processing.
// See PRODUCTION_PLAN.md Phase 5
type DocumentHandler struct {
	invoiceService *service.InvoiceService
}

// NewDocumentHandler creates a new DocumentHandler.
func NewDocumentHandler(s *service.InvoiceService) *DocumentHandler {
	return &DocumentHandler{invoiceService: s}
}

// AnalyzeDocumentRequest is the payload for document analysis.
type AnalyzeDocumentRequest struct {
	DocumentID uuid.UUID `json:"document_id"`
}

// AnalyzeDocument handles POST /api/v1/documents/analyze
// See API_AND_TYPES_SPEC.md Section 3.1
func (h *DocumentHandler) AnalyzeDocument(w http.ResponseWriter, r *http.Request) {
	// 1. Multi-tenancy gate: Extract X-Org-ID header
	orgIDStr := r.Header.Get("X-Org-ID")
	if orgIDStr == "" {
		http.Error(w, "X-Org-ID header is required", http.StatusBadRequest)
		return
	}
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "Invalid X-Org-ID", http.StatusBadRequest)
		return
	}

	// 2. Parse request body
	var req AnalyzeDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.DocumentID == uuid.Nil {
		http.Error(w, "document_id is required", http.StatusBadRequest)
		return
	}

	// 3. Call AnalyzeInvoice
	projectID, extraction, err := h.invoiceService.AnalyzeInvoice(r.Context(), orgID, req.DocumentID)
	if err != nil {
		// Log error and return generic message for security
		// In production, we'd distinguish between 404 (not found) and 500 (ai error)
		http.Error(w, "Analysis failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Persist result
	_, err = h.invoiceService.SaveExtraction(r.Context(), projectID, extraction)
	if err != nil {
		http.Error(w, "Failed to save extraction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(extraction)
}
