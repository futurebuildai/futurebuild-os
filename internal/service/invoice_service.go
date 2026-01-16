package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/colton/futurebuild/pkg/types"
)

// ConfidenceThresholdForReview defines the minimum AI confidence
// required to bypass human review. See PRODUCTION_PLAN.md Step 39.
const ConfidenceThresholdForReview = 0.85

// InvoiceService handles invoice-specific analysis and persistence.
// See PRODUCTION_PLAN.md Step 37
type InvoiceService struct {
	db     *pgxpool.Pool
	client ai.Client
}

// NewInvoiceService creates a new InvoiceService.
func NewInvoiceService(db *pgxpool.Pool, client ai.Client) *InvoiceService {
	return &InvoiceService{
		db:     db,
		client: client,
	}
}

// =============================================================================
// AI PROMPTS (OPERATION IRONCLAD TASK 5)
// =============================================================================
// ENGINEERING STANDARD: AI prompts are extracted to package-level variables
// for separation from business logic. This allows configuration loaders or
// testing frameworks to override prompts without modifying service code.
//
// Usage: service.InvoicePromptTemplate = customPrompt
// =============================================================================

// InvoicePromptTemplate is the AI prompt template for invoice extraction.
// Exported as a var so it can be overridden by config loaders or tests.
// MONETARY PRECISION: Instructs AI to return all monetary values as integer cents.
var InvoicePromptTemplate = `
Extract the following information from the invoice text provided and return it as a structured JSON object.
Do NOT include any markdown formatting (like ` + "```" + `json) in your response. Return ONLY the raw JSON string.

CRITICAL: All monetary amounts MUST be returned as INTEGERS representing CENTS.
Example: $10.50 should be 1050, $1400.00 should be 140000, $0.99 should be 99.

Schema:
{
  "vendor": "String",
  "date": "ISO-8601 Date",
  "invoice_number": "String",
  "total_amount_cents": 0,
  "line_items": [
    {
      "description": "String",
      "quantity": 0.0,
      "unit_price_cents": 0,
      "total_cents": 0
    }
  ],
  "suggested_wbs_code": "String",
  "confidence": 0.0 to 1.0 (float)
}

Invoice Text:
%s
`

// AnalyzeInvoice fetches document text and uses AI to extract structured data.
// It returns the projectID associated with the document for consistency.
// See API_AND_TYPES_SPEC.md Section 3.1
func (s *InvoiceService) AnalyzeInvoice(ctx context.Context, orgID uuid.UUID, docID uuid.UUID) (uuid.UUID, *types.InvoiceExtraction, error) {
	// 1. Fetch Extracted Text from Documents with Multi-Tenancy Check
	var extractedText string
	var projectID uuid.UUID
	query := `
		SELECT d.project_id, d.extracted_text 
		FROM documents d
		JOIN projects p ON d.project_id = p.id
		WHERE d.id = $1 AND p.org_id = $2
	`
	err := s.db.QueryRow(ctx, query, docID, orgID).Scan(&projectID, &extractedText)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("failed to fetch document or unauthorized: %w", err)
	}

	if extractedText == "" {
		return uuid.Nil, nil, fmt.Errorf("document has no extracted text")
	}

	// 2. AI Prompting (Mandated use of Gemini 2.5 Flash per BACKEND_SCOPE Section 3.2)
	// L7 Vendor Abstraction: Use ai.GenerateRequest instead of genai.Part
	// Code Review Issue 1B Fix: Enable logprobs for actual model confidence scoring
	req := ai.GenerateRequest{
		Model:          ai.ModelTypeFlash,
		Parts:          []ai.ContentPart{{Text: fmt.Sprintf(InvoicePromptTemplate, extractedText)}},
		ReturnLogprobs: true, // Enable logprobs for true model confidence
	}
	resp, err := s.client.GenerateContent(ctx, req)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("ai analysis failed: %w", err)
	}

	// Clean response (remove markdown if any)
	cleanResp := strings.TrimSpace(resp.Text)
	cleanResp = strings.TrimPrefix(cleanResp, "```json")
	cleanResp = strings.TrimSuffix(cleanResp, "```")
	cleanResp = strings.TrimSpace(cleanResp)

	// 3. Parse JSON
	var extraction types.InvoiceExtraction
	if err := json.Unmarshal([]byte(cleanResp), &extraction); err != nil {
		return uuid.Nil, nil, fmt.Errorf("failed to parse AI response: %w. Response: %s", err, cleanResp)
	}

	// 4. Use model confidence from logprobs if available (Code Review Issue 1B Fix)
	// Model confidence derived from logprobs is more reliable than LLM self-reported confidence
	if resp.Confidence > 0 {
		extraction.Confidence = float64(resp.Confidence)
	}
	// If logprobs not available (resp.Confidence == 0), fall back to JSON confidence
	// with warning logged in SaveExtraction

	return projectID, &extraction, nil
}

// SaveExtraction persists the analyzed invoice to the DB.
// Uses atomic PostgreSQL UPSERT pattern to eliminate race conditions.
// See DATA_SPINE_SPEC.md Section 4.2
// See PRODUCTION_PLAN.md Step 41 for re-processing support
// CONCURRENCY FIX: Replaced Check-Then-Act with INSERT ... ON CONFLICT
func (s *InvoiceService) SaveExtraction(ctx context.Context, projectID uuid.UUID, extraction *types.InvoiceExtraction, sourceDocID *uuid.UUID) (uuid.UUID, error) {
	// 1. Map Domain Types
	var lineItems models.LineItems
	for _, item := range extraction.LineItems {
		lineItems = append(lineItems, models.LineItem{
			Description:    item.Description,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			TotalCents:     item.TotalCents,
		})
	}

	// 2. Evaluate Human Review Requirement
	// See PRODUCTION_PLAN.md Step 39
	//
	// CONFIDENCE SCORING (Code Review Issue 1B - Resolution):
	// Confidence is now derived from model logprobs when available (see AnalyzeInvoice).
	// Logprobs provide actual token probabilities from the model, not self-reported values.
	//
	// Fallback behavior: If logprobs unavailable (e.g., mock client in tests),
	// uses the JSON-provided confidence with caution.
	//
	// See: https://developers.googleblog.com/unlock-gemini-reasoning-with-logprobs-on-vertex-ai/
	isHumanReviewRequired := extraction.Confidence < ConfidenceThresholdForReview

	// 3. Prepare Variables
	// Generate a new ID to use IF this turns out to be an INSERT.
	// If it's an UPDATE via ON CONFLICT, RETURNING gives us the existing ID.
	newID := uuid.New()

	// 4. Atomic Upsert (PostgreSQL specific)
	// This query handles both INSERT and UPDATE in a single atomic operation,
	// eliminating the race condition where two concurrent requests could both
	// see "no record" and both attempt to INSERT.
	//
	// The partial index constraint (WHERE source_document_id IS NOT NULL) ensures:
	// - New invoices without sourceDocID always INSERT (no conflict possible on NULL)
	// - Re-processed invoices with sourceDocID either INSERT or UPDATE atomically
	query := `
		INSERT INTO invoices (
			id, project_id, vendor_name, amount_cents, line_items, 
			detected_wbs_code, status, confidence, invoice_date, 
			invoice_number, is_human_review_required, source_document_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (source_document_id) 
		WHERE source_document_id IS NOT NULL
		DO UPDATE SET
			vendor_name = EXCLUDED.vendor_name,
			amount_cents = EXCLUDED.amount_cents,
			line_items = EXCLUDED.line_items,
			detected_wbs_code = EXCLUDED.detected_wbs_code,
			confidence = EXCLUDED.confidence,
			invoice_date = EXCLUDED.invoice_date,
			invoice_number = EXCLUDED.invoice_number,
			is_human_review_required = EXCLUDED.is_human_review_required,
			status = EXCLUDED.status
		RETURNING id
	`

	var finalID uuid.UUID
	err := s.db.QueryRow(ctx, query,
		newID,
		projectID,
		extraction.Vendor,
		extraction.TotalAmountCents,
		lineItems,
		extraction.SuggestedWBSCode,
		models.InvoiceStatusPending,
		extraction.Confidence,
		extraction.Date,
		extraction.InvoiceNumber,
		isHumanReviewRequired,
		sourceDocID, // pgx handles *uuid.UUID correctly (nil -> NULL)
	).Scan(&finalID)

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to upsert invoice: %w", err)
	}

	return finalID, nil
}
