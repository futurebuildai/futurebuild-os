package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/colton/futurebuild/internal/audit"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/prompts"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// VisionService implements AI-powered site photo analysis.
// See API_AND_TYPES_SPEC.md Section 2.2 and PRODUCTION_PLAN.md Step 40.
type VisionService struct {
	client     ai.Client
	httpClient *http.Client
	logger     audit.AgentLogger
}

// NewVisionService creates a new VisionService.
func NewVisionService(client ai.Client, logger audit.AgentLogger) *VisionService {
	return &VisionService{
		client: client,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// confidenceWarningThreshold is the threshold below which a field is flagged
// with a warning for user verification. Matches frontend's < 0.8 threshold
// in onboarding-store.ts fieldsNeedingVerification.
const confidenceWarningThreshold = 0.8

// ParseDocument extracts structured construction data from uploaded plan files.
// Sprint 2.1 Task 2.1.1: Standalone extraction endpoint for the Vision Pipeline.
// Returns a VisionExtractionResponse with extracted values, ConfidenceReport, and long-lead items.
func (s *VisionService) ParseDocument(ctx context.Context, fileBytes []byte, mimeType string) (*models.VisionExtractionResponse, error) {
	startTime := time.Now()

	tracer := otel.Tracer("futurebuild.vision")
	ctx, span := tracer.Start(ctx, "Vision.ParseDocument")
	defer span.End()
	span.SetAttributes(attribute.String("mime.type", mimeType))

	prompt := prompts.BlueprintExtractionPrompt()

	slog.Info("vision: parsing document",
		"mime_type", mimeType,
		"file_size", len(fileBytes),
	)

	// Create multimodal request with the uploaded file
	req := ai.NewMultimodalRequest(ai.ModelTypeFlash, prompt, fileBytes, mimeType)
	req.ReturnLogprobs = true

	result, err := s.client.GenerateContent(ctx, req)
	if err != nil {
		slog.Error("vision: extraction failed, falling back to manual mode", "error", err.Error())
		return &models.VisionExtractionResponse{
			ExtractedValues: make(map[string]any),
			ConfidenceReport: models.ConfidenceReport{
				OverallConfidence: 0.0,
				FieldConfidences:  make(map[string]float64),
				Warnings:          []string{"AI service unavailable. Please enter details manually."},
			},
		}, nil
	}

	// Strip markdown code block wrappers if present
	jsonText := strings.TrimSpace(result.Text)
	jsonText = strings.TrimPrefix(jsonText, "```json")
	jsonText = strings.TrimPrefix(jsonText, "```")
	jsonText = strings.TrimSuffix(jsonText, "```")
	jsonText = strings.TrimSpace(jsonText)

	// Parse JSON response (same structure as BlueprintExtractionPrompt output)
	var extraction struct {
		Name           string  `json:"name"`
		Address        string  `json:"address"`
		SquareFootage  float64 `json:"square_footage"`
		FoundationType string  `json:"foundation_type"`
		Stories        int     `json:"stories"`
		Bedrooms       int     `json:"bedrooms"`
		Bathrooms      int     `json:"bathrooms"`
		LongLeadItems  []struct {
			Name     string `json:"name"`
			Brand    string `json:"brand"`
			Model    string `json:"model"`
			Category string `json:"category"`
			Notes    string `json:"notes"`
		} `json:"long_lead_items"`
		Confidence map[string]float64 `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(jsonText), &extraction); err != nil {
		return nil, fmt.Errorf("failed to parse extraction result: %w", err)
	}

	// Build extracted values map (only non-empty fields)
	values := make(map[string]any)
	if extraction.Name != "" {
		values["name"] = extraction.Name
	}
	if extraction.Address != "" {
		values["address"] = extraction.Address
	}
	if extraction.SquareFootage > 0 {
		values["square_footage"] = extraction.SquareFootage
	}
	if extraction.FoundationType != "" {
		values["foundation_type"] = extraction.FoundationType
	}
	if extraction.Stories > 0 {
		values["stories"] = extraction.Stories
	}
	if extraction.Bedrooms > 0 {
		values["bedrooms"] = extraction.Bedrooms
	}
	if extraction.Bathrooms > 0 {
		values["bathrooms"] = extraction.Bathrooms
	}

	// Compute ConfidenceReport
	fieldConfidences := extraction.Confidence
	if fieldConfidences == nil {
		fieldConfidences = make(map[string]float64)
	}

	var warnings []string
	var suggestedQuestions []string

	// Generate warnings and questions for low-confidence or missing fields
	priorityFields := models.GetPriorityFields()
	for _, pf := range priorityFields {
		conf, hasConf := fieldConfidences[pf.Field]
		_, hasValue := values[pf.Field]

		if !hasValue {
			warnings = append(warnings, fmt.Sprintf("Field '%s' could not be extracted from the document", pf.Field))
			suggestedQuestions = append(suggestedQuestions, pf.Question)
		} else if hasConf && conf < confidenceWarningThreshold {
			warnings = append(warnings, fmt.Sprintf("Field '%s' has low confidence (%.0f%%)", pf.Field, conf*100))
			suggestedQuestions = append(suggestedQuestions, pf.Question)
		}
	}

	// Calculate overall confidence
	overallConfidence := 0.0
	if len(fieldConfidences) > 0 {
		sum := 0.0
		for _, c := range fieldConfidences {
			sum += c
		}
		overallConfidence = sum / float64(len(fieldConfidences))
	}

	// Enrich long-lead items with lead time estimates
	var longLeadItems []models.LongLeadItem
	if len(extraction.LongLeadItems) > 0 {
		leadTimes := models.KnownBrandLeadTimes()
		for _, item := range extraction.LongLeadItems {
			leadWeeks := estimateLeadTimeForVision(item.Brand, item.Category, leadTimes)
			longLeadItems = append(longLeadItems, models.LongLeadItem{
				Name:               item.Name,
				Brand:              item.Brand,
				Model:              item.Model,
				Category:           item.Category,
				EstimatedLeadWeeks: leadWeeks,
				Notes:              item.Notes,
			})
		}
	}

	slog.Info("vision: extraction completed",
		"extracted_fields", len(values),
		"overall_confidence", overallConfidence,
		"warnings", len(warnings),
		"long_lead_items", len(longLeadItems),
	)

	if s.logger != nil {
		_ = s.logger.LogDecision(context.Background(), audit.AgentDecisionEntry{
			Timestamp:    time.Now(),
			Agent:        "VisionAgent",
			Action:       "ParseDocument",
			InputSummary: fmt.Sprintf("Blueprint (size: %d, mime: %s)", len(fileBytes), mimeType),
			Decision:     fmt.Sprintf("Extracted %d fields and %d long-lead items", len(values), len(longLeadItems)),
			Confidence:   overallConfidence,
			Model:        string(ai.ModelTypeFlash),
			LatencyMS:    time.Since(startTime).Milliseconds(),
			ProjectID:    "", // typically project creation happens after extraction
			UserID:       "", // from context if available, otherwise blank
			TraceID:      span.SpanContext().TraceID().String(),
		})
	}

	return &models.VisionExtractionResponse{
		ExtractedValues: values,
		ConfidenceReport: models.ConfidenceReport{
			OverallConfidence:  overallConfidence,
			FieldConfidences:   fieldConfidences,
			Warnings:           warnings,
			SuggestedQuestions: suggestedQuestions,
		},
		LongLeadItems: longLeadItems,
		RawText:       result.Text,
	}, nil
}

// estimateLeadTimeForVision determines lead time for vision-extracted items.
func estimateLeadTimeForVision(brand, category string, leadTimes map[string]int) int {
	brandLower := strings.ToLower(strings.TrimSpace(brand))
	if weeks, ok := leadTimes[brandLower]; ok {
		return weeks
	}
	for key, weeks := range leadTimes {
		if strings.Contains(brandLower, key) || strings.Contains(key, brandLower) {
			return weeks
		}
	}
	categoryDefaults := map[string]int{
		"windows": 8, "doors": 6, "hvac": 4,
		"appliances": 6, "millwork": 8, "finishes": 4,
	}
	if weeks, ok := categoryDefaults[strings.ToLower(category)]; ok {
		return weeks
	}
	return 4
}

const visionPromptTemplate = `
Analyze the provided image and determine if the following task has been completed or is accurately represented:
Task Description: %s

Return your analysis as a structured JSON object with the following schema:
{
  "is_verified": bool,
  "confidence": 0.0 to 1.0 (float),
  "reasoning": "Brief explanation of what was observed"
}

Do NOT include any markdown formatting (like ` + "```" + `json) in your response. Return ONLY the raw JSON string.
`

// VerifyTask returns (is_verified, confidence_score, error)
// See API_AND_TYPES_SPEC.md Section 2.2
func (s *VisionService) VerifyTask(ctx context.Context, imageURL string, taskDescription string) (bool, float64, error) {
	// 1. Download image from URL
	imageBytes, mimeType, err := s.downloadImage(ctx, imageURL)
	if err != nil {
		return false, 0, fmt.Errorf("failed to download image: %w", err)
	}

	// 2. Prepare AI Request using vendor-agnostic types
	// L7 Vendor Abstraction: Use ai.NewMultimodalRequest instead of genai.Part
	req := ai.NewMultimodalRequest(
		ai.ModelTypeFlash,
		fmt.Sprintf(visionPromptTemplate, taskDescription),
		imageBytes,
		mimeType,
	)

	// 3. Call AI (vendor-agnostic)
	resp, err := s.client.GenerateContent(ctx, req)
	if err != nil {
		return false, 0, fmt.Errorf("ai vision analysis failed: %w", err)
	}

	// Clean response
	cleanResp := strings.TrimSpace(resp.Text)
	cleanResp = strings.TrimPrefix(cleanResp, "```json")
	cleanResp = strings.TrimSuffix(cleanResp, "```")
	cleanResp = strings.TrimSpace(cleanResp)

	// 4. Parse Result
	var result struct {
		IsVerified bool    `json:"is_verified"`
		Confidence float64 `json:"confidence"`
		Reasoning  string  `json:"reasoning"`
	}
	if err := json.Unmarshal([]byte(cleanResp), &result); err != nil {
		return false, 0, fmt.Errorf("failed to parse vision response: %w. Response: %s", err, cleanResp)
	}

	return result.IsVerified, result.Confidence, nil
}

// VerifyAndPersistTask verifies a task via site photo AND persists the result.
// This satisfies the "Database is State" requirement from DATA_SPINE_SPEC.md.
// See PRODUCTION_PLAN.md Step 40
func (s *VisionService) VerifyAndPersistTask(ctx context.Context, db DBExecutor, taskID, projectID, orgID uuid.UUID, imageURL string, taskDescription string) (bool, float64, error) {
	startTime := time.Now()

	tracer := otel.Tracer("futurebuild.vision")
	ctx, span := tracer.Start(ctx, "Vision.VerifyAndPersistTask")
	defer span.End()
	span.SetAttributes(
		attribute.String("project.id", projectID.String()),
		attribute.String("task.id", taskID.String()),
		attribute.String("org.id", orgID.String()),
	)

	// 1. Call AI Verification
	isVerified, confidence, err := s.VerifyTask(ctx, imageURL, taskDescription)
	if err != nil {
		return false, 0, fmt.Errorf("vision verification failed: %w", err)
	}

	// 2. Persist to Database (with Multi-Tenancy Check)
	// See DATA_SPINE_SPEC.md Section 3.3
	query := `
		UPDATE project_tasks pt
		SET 
			verified_by_vision = $1,
			verification_confidence = $2,
			updated_at = NOW()
		FROM projects p
		WHERE pt.id = $3 
			AND pt.project_id = $4
			AND p.id = $4
			AND p.org_id = $5
	`
	result, err := db.Exec(ctx, query, isVerified, confidence, taskID, projectID, orgID)
	if err != nil {
		return false, 0, fmt.Errorf("failed to persist verification result: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return false, 0, fmt.Errorf("task not found or access denied (multi-tenancy violation)")
	}

	if s.logger != nil {
		decision := "Task verified as false"
		if isVerified {
			decision = "Task verified as true"
		}
		_ = s.logger.LogDecision(context.Background(), audit.AgentDecisionEntry{
			Timestamp:    time.Now(),
			Agent:        "VisionAgent",
			Action:       "VerifyTask",
			InputSummary: fmt.Sprintf("Photo upload vs task: %s", taskDescription),
			Decision:     decision,
			Confidence:   confidence,
			Model:        string(ai.ModelTypeFlash),
			LatencyMS:    time.Since(startTime).Milliseconds(),
			ProjectID:    projectID.String(),
			UserID:       "",
			TraceID:      span.SpanContext().TraceID().String(),
		})
	}

	return isVerified, confidence, nil
}

// DBExecutor defines the minimal database interface needed for persistence.
// This allows both *pgxpool.Pool and *pgx.Tx to be used.
type DBExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

func (s *VisionService) downloadImage(ctx context.Context, url string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	mimeType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(mimeType, "image/") {
		return nil, "", fmt.Errorf("invalid content type: %s", mimeType)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return data, mimeType, nil
}

// ProgressClassification is the result of classifying a construction progress photo.
type ProgressClassification struct {
	DetectedPhase    string   `json:"detected_phase"`
	WBSCode          string   `json:"wbs_code"`
	Confidence       float64  `json:"confidence"`
	VisibleElements  []string `json:"visible_elements"`
	EstimatedPercent int      `json:"estimated_percent"`
	Recommendations  []string `json:"recommendations"`
}

// ClassifyProgressPhoto analyzes a construction site photo and classifies it
// to a WBS phase with an estimated completion percentage.
func (s *VisionService) ClassifyProgressPhoto(ctx context.Context, imageBytes []byte, mimeType string) (*ProgressClassification, error) {
	if s.client == nil {
		return nil, fmt.Errorf("vision client not configured")
	}

	resp, err := s.client.GenerateContent(ctx, ai.GenerateRequest{
		Model: ai.ModelTypeFlash,
		Parts: []ai.ContentPart{
			{Text: prompts.ProgressPhotoClassificationPrompt},
			{Data: imageBytes, MimeType: mimeType},
		},
		Temperature: 0.2,
		MaxTokens:   1024,
	})
	if err != nil {
		return nil, fmt.Errorf("vision classify: %w", err)
	}

	if resp.Text == "" {
		return nil, fmt.Errorf("empty response from vision model")
	}

	// Parse structured JSON response
	var classification ProgressClassification
	text := resp.Text
	// Strip markdown code fences if present
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	if err := json.Unmarshal([]byte(text), &classification); err != nil {
		slog.Warn("vision classify: failed to parse structured response, attempting best-effort",
			"raw", text, "error", err)
		return nil, fmt.Errorf("parse classification response: %w", err)
	}

	return &classification, nil
}

