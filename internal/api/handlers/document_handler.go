package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// DocumentHandler handles requests related to document processing.
// See PRODUCTION_PLAN.md Phase 5
type DocumentHandler struct {
	invoiceService  service.InvoiceServicer
	documentService service.DocumentServicer // See PRODUCTION_PLAN.md Step 41
}

// NewDocumentHandler creates a new DocumentHandler.
func NewDocumentHandler(invoiceSvc service.InvoiceServicer, docSvc service.DocumentServicer) *DocumentHandler {
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
// L7 Security Fix: Uses JWT claims for multi-tenancy, NOT headers (Confused Deputy prevention)
func (h *DocumentHandler) AnalyzeDocument(w http.ResponseWriter, r *http.Request) {
	// 1. Multi-tenancy gate: Extract OrgID from validated JWT claims
	// See PRODUCTION_PLAN.md Step 61.1 (Confused Deputy Remediation)
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		slog.Warn("doc: missing authentication context", "method", r.Method, "error", err)
		http.Error(w, "Unauthorized: missing authentication", http.StatusUnauthorized)
		return
	}
	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		slog.Warn("doc: invalid OrgID in JWT claims", "raw_org_id", claims.OrgID, "error", err)
		http.Error(w, "Unauthorized: invalid credentials", http.StatusUnauthorized)
		return
	}

	// 2. Parse request body (L7: limit body size to prevent DoS)
	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var req AnalyzeDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("doc: invalid request payload", "error", err, "org_id", orgID)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.DocumentID == uuid.Nil {
		slog.Warn("doc: document_id is required", "org_id", orgID)
		http.Error(w, "document_id is required", http.StatusBadRequest)
		return
	}

	slog.Info("doc: analyzing document", "document_id", req.DocumentID, "org_id", orgID)

	// 3. Call AnalyzeInvoice
	projectID, extraction, err := h.invoiceService.AnalyzeInvoice(r.Context(), orgID, req.DocumentID)
	if err != nil {
		slog.Error("doc: analysis failed", "document_id", req.DocumentID, "org_id", orgID, "error", err)
		http.Error(w, "Analysis failed", http.StatusInternalServerError)
		return
	}

	// 4. Persist result
	// See PRODUCTION_PLAN.md Step 41: Always pass sourceDocID to enable UPSERT/Idempotency
	_, err = h.invoiceService.SaveExtraction(r.Context(), projectID, extraction, &req.DocumentID)
	if err != nil {
		slog.Error("doc: failed to save extraction", "document_id", req.DocumentID, "project_id", projectID, "error", err)
		http.Error(w, "Failed to save extraction", http.StatusInternalServerError)
		return
	}

	slog.Info("doc: analysis completed", "document_id", req.DocumentID, "project_id", projectID, "vendor", extraction.Vendor)

	// 5. Return JSON
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(extraction)
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
// L7 Security Fix: Uses JWT claims for multi-tenancy, NOT headers (Confused Deputy prevention)
func (h *DocumentHandler) ReprocessDocument(w http.ResponseWriter, r *http.Request) {
	// 1. Multi-tenancy gate: Extract OrgID from validated JWT claims
	// See PRODUCTION_PLAN.md Step 61.1 (Confused Deputy Remediation)
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		slog.Warn("doc: missing authentication context for reprocess", "method", r.Method, "error", err)
		http.Error(w, "Unauthorized: missing authentication", http.StatusUnauthorized)
		return
	}
	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		slog.Warn("doc: invalid OrgID in JWT claims for reprocess", "raw_org_id", claims.OrgID, "error", err)
		http.Error(w, "Unauthorized: invalid credentials", http.StatusUnauthorized)
		return
	}

	// 2. Extract document ID from URL path
	docIDStr := chi.URLParam(r, "id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		slog.Warn("doc: invalid document ID for reprocess", "raw_doc_id", docIDStr, "error", err)
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	slog.Info("doc: reprocessing document", "document_id", docID, "org_id", orgID)

	// 3. Call ReprocessDocument
	err = h.documentService.ReprocessDocument(r.Context(), orgID, docID)
	if err != nil {
		slog.Error("doc: reprocess failed", "document_id", docID, "org_id", orgID, "error", err)
		http.Error(w, "Reprocess failed", http.StatusInternalServerError)
		return
	}

	// 4. Trigger Invoice Re-extraction (The "Missing Chain" from CTO Audit Finding C2)
	projectID, extraction, err := h.invoiceService.AnalyzeInvoice(r.Context(), orgID, docID)
	if err != nil {
		slog.Error("doc: re-extraction failed", "document_id", docID, "org_id", orgID, "error", err)
		http.Error(w, "Re-extraction failed", http.StatusInternalServerError)
		return
	}

	// 5. Persist the re-extracted invoice (Uses UPSERT via sourceDocID)
	_, err = h.invoiceService.SaveExtraction(r.Context(), projectID, extraction, &docID)
	if err != nil {
		slog.Error("doc: failed to save re-extraction", "document_id", docID, "project_id", projectID, "error", err)
		http.Error(w, "Failed to save re-extraction", http.StatusInternalServerError)
		return
	}

	// 6. Fetch updated document status for response (Replaces Raw SQL - CTO Finding A2)
	processingStatus, reprocessedCount, err := h.documentService.GetDocumentStatus(r.Context(), docID)
	if err != nil {
		// Non-fatal: return success with defaults
		slog.Warn("doc: failed to get document status (non-fatal)", "document_id", docID, "error", err)
		processingStatus = "completed"
		reprocessedCount = 0
	}

	slog.Info("doc: reprocessing completed", "document_id", docID, "status", processingStatus, "reprocessed_count", reprocessedCount)

	// 5. Return JSON response
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ReprocessDocumentResponse{
		DocumentID:       docID,
		Status:           processingStatus,
		ReprocessedCount: reprocessedCount,
	})
}
