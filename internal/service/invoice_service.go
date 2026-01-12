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
	"google.golang.org/genai"
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

const invoicePromptTemplate = `
Extract the following information from the invoice text provided and return it as a structured JSON object.
Do NOT include any markdown formatting (like ` + "```" + `json) in your response. Return ONLY the raw JSON string.

Schema:
{
  "vendor": "String",
  "date": "ISO-8601 Date",
  "invoice_number": "String",
  "total_amount": 0.00,
  "line_items": [
    {
      "description": "String",
      "quantity": 0.0,
      "unit_price": 0.0,
      "total": 0.0
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
	// 2. AI Prompting (Mandated use of Gemini 2.5 Flash per BACKEND_SCOPE Section 3.2)
	promptPart := &genai.Part{Text: fmt.Sprintf(invoicePromptTemplate, extractedText)}
	resp, err := s.client.GenerateContent(ctx, ai.ModelTypeFlash, promptPart)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("ai analysis failed: %w", err)
	}

	// Clean response (remove markdown if any)
	cleanResp := strings.TrimSpace(resp)
	cleanResp = strings.TrimPrefix(cleanResp, "```json")
	cleanResp = strings.TrimSuffix(cleanResp, "```")
	cleanResp = strings.TrimSpace(cleanResp)

	// 3. Parse JSON
	var extraction types.InvoiceExtraction
	if err := json.Unmarshal([]byte(cleanResp), &extraction); err != nil {
		return uuid.Nil, nil, fmt.Errorf("failed to parse AI response: %w. Response: %s", err, cleanResp)
	}

	return projectID, &extraction, nil
}

// SaveExtraction persists the analyzed invoice to the DB.
// See DATA_SPINE_SPEC.md Section 4.2
func (s *InvoiceService) SaveExtraction(ctx context.Context, projectID uuid.UUID, extraction *types.InvoiceExtraction) (uuid.UUID, error) {
	invoiceID := uuid.New()

	// Map types.InvoiceExtraction to models.LineItems
	var lineItems models.LineItems
	for _, item := range extraction.LineItems {
		lineItems = append(lineItems, models.LineItem{
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Total:       item.Total,
		})
	}

	// Logic Integration: Flag for human review if confidence is low
	// See PRODUCTION_PLAN.md Step 39
	isHumanReviewRequired := extraction.Confidence < ConfidenceThresholdForReview

	query := `
		INSERT INTO invoices (
			id, project_id, vendor_name, amount, line_items, 
			detected_wbs_code, status, confidence, invoice_date, 
			invoice_number, is_human_review_required
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := s.db.Exec(ctx, query,
		invoiceID,
		projectID,
		extraction.Vendor,
		extraction.TotalAmount,
		lineItems,
		extraction.SuggestedWBSCode,
		models.InvoiceStatusPending,
		extraction.Confidence,
		extraction.Date,
		extraction.InvoiceNumber,
		isHumanReviewRequired,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to save invoice: %w", err)
	}

	return invoiceID, nil
}
