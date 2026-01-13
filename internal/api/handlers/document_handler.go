package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/colton/futurebuild/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// DocumentHandler handles requests related to document processing.
// See PRODUCTION_PLAN.md Phase 5
type DocumentHandler struct {
	invoiceService  *service.InvoiceService
	documentService *service.DocumentService // See PRODUCTION_PLAN.md Step 41
}

// NewDocumentHandler creates a new DocumentHandler.
func NewDocumentHandler(invoiceSvc *service.InvoiceService, docSvc *service.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		invoiceService:  invoiceSvc,
		documentService: docSvc,
	}
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
	// See PRODUCTION_PLAN.md Step 41: Always pass sourceDocID to enable UPSERT/Idempotency
	_, err = h.invoiceService.SaveExtraction(r.Context(), projectID, extraction, &req.DocumentID)
	if err != nil {
		http.Error(w, "Failed to save extraction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(extraction)
}

// ReprocessDocumentResponse is the response for document re-processing.
type ReprocessDocumentResponse struct {
	DocumentID       uuid.UUID `json:"document_id"`
	Status           string    `json:"status"`
	ReprocessedCount int       `json:"reprocessed_count"`
}

// ReprocessDocument handles POST /api/v1/documents/{id}/reprocess
// Triggers re-analysis of a document with updated content.
// See PRODUCTION_PLAN.md Step 41
func (h *DocumentHandler) ReprocessDocument(w http.ResponseWriter, r *http.Request) {
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

	// 2. Extract document ID from URL path
	docIDStr := chi.URLParam(r, "id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	// 3. Call ReprocessDocument
	err = h.documentService.ReprocessDocument(r.Context(), orgID, docID)
	if err != nil {
		http.Error(w, "Reprocess failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Trigger Invoice Re-extraction (The "Missing Chain" from CTO Audit Finding C2)
	projectID, extraction, err := h.invoiceService.AnalyzeInvoice(r.Context(), orgID, docID)
	if err != nil {
		http.Error(w, "Re-extraction failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Persist the re-extracted invoice (Uses UPSERT via sourceDocID)
	_, err = h.invoiceService.SaveExtraction(r.Context(), projectID, extraction, &docID)
	if err != nil {
		http.Error(w, "Failed to save re-extraction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 6. Fetch updated document status for response (Replaces Raw SQL - CTO Finding A2)
	processingStatus, reprocessedCount, err := h.documentService.GetDocumentStatus(r.Context(), docID)
	if err != nil {
		// Non-fatal: return success with defaults
		processingStatus = "completed"
		reprocessedCount = 0
	}

	// 5. Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ReprocessDocumentResponse{
		DocumentID:       docID,
		Status:           processingStatus,
		ReprocessedCount: reprocessedCount,
	})
}
