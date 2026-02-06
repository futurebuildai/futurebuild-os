package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/prompts"
	"github.com/colton/futurebuild/pkg/ai"
)

// InterrogatorService implements the onboarding agent that extracts project data
// from user conversations and documents.
type InterrogatorService struct {
	aiClient ai.Client
}

// NewInterrogatorService creates a new interrogator service.
func NewInterrogatorService(aiClient ai.Client) *InterrogatorService {
	return &InterrogatorService{
		aiClient: aiClient,
	}
}

// ProcessMessage handles a single turn of the onboarding conversation.
func (s *InterrogatorService) ProcessMessage(
	ctx context.Context,
	userID, tenantID string,
	req *models.OnboardRequest,
) (*models.OnboardResponse, error) {
	// L7: Structured logging for observability
	slog.Info("onboarding_message_received",
		"user_id", userID,
		"tenant_id", tenantID,
		"session_id", req.SessionID,
		"has_message", req.Message != "",
		"has_document", req.DocumentURL != "",
		"has_file_data", len(req.DocumentData) > 0,
		"current_fields", len(req.CurrentState),
	)

	resp := &models.OnboardResponse{
		SessionID:        req.SessionID,
		ExtractedValues:  make(map[string]any),
		ConfidenceScores: make(map[string]float64),
		ReadyToCreate:    false,
	}

	// BRANCH 1: Document uploaded → Extract via Vision API
	// Step 77: Support both URL-based and inline file data paths.
	if req.DocumentURL != "" || len(req.DocumentData) > 0 {
		var extraction *models.ExtractionResult
		var err error

		if len(req.DocumentData) > 0 {
			// Step 77: Inline file data (multipart upload)
			extraction, err = s.extractFromBytes(ctx, req.DocumentData, req.DocumentContentType, req.DocumentFileName)
		} else {
			// Existing URL-based path
			extraction, err = s.extractFromDocument(ctx, req.DocumentURL)
		}

		if err != nil {
			resp.Reply = "I couldn't read that file. Could you try a clearer scan or describe the project?"
			return resp, nil
		}

		// Merge extracted values into response
		for k, v := range extraction.Values {
			resp.ExtractedValues[k] = v
			resp.ConfidenceScores[k] = extraction.Confidence[k]
		}

		// Include long-lead items in response
		resp.LongLeadItems = extraction.LongLeadItems

		// Generate a summary message with procurement warnings
		resp.Reply = s.generateExtractionSummary(extraction)
	}

	// BRANCH 2: User message → Parse natural language
	if req.Message != "" {
		extraction, err := s.parseUserMessage(ctx, req.Message, req.CurrentState)
		if err == nil {
			for k, v := range extraction.Values {
				resp.ExtractedValues[k] = v
				resp.ConfidenceScores[k] = extraction.Confidence[k]
			}
		}
	}

	// Merge with existing state
	mergedState := s.mergeStates(req.CurrentState, resp.ExtractedValues)

	// Check if ready to create (name + address are required)
	resp.ReadyToCreate = s.checkReadyToCreate(mergedState)

	// Generate next question if not ready
	if !resp.ReadyToCreate {
		nextField, question := s.getNextQuestion(mergedState)
		resp.NextPriorityField = nextField
		resp.ClarifyingQuestion = question

		// If no explicit reply yet, use the clarifying question
		if resp.Reply == "" {
			resp.Reply = question
		}
	} else if resp.Reply == "" {
		resp.Reply = "Great! Your project is ready to create. Review the details and click 'Create Project' when ready."
	}

	// L7: Log completion metrics
	avgConfidence := calculateAvgConfidence(resp.ConfidenceScores)
	slog.Info("onboarding_message_completed",
		"session_id", req.SessionID,
		"extracted_fields", len(resp.ExtractedValues),
		"avg_confidence", avgConfidence,
		"ready_to_create", resp.ReadyToCreate,
		"next_field", resp.NextPriorityField,
	)

	return resp, nil
}

// extractFromDocument uses Vision API to extract structured data from blueprints.
// L7: Uses secure image download with SSRF protection.
func (s *InterrogatorService) extractFromDocument(
	ctx context.Context,
	documentURL string,
) (*models.ExtractionResult, error) {
	prompt := prompts.BlueprintExtractionPrompt()

	// SECURITY: Download image with SSRF protection
	imageData, mimeType, err := s.downloadImage(ctx, documentURL)
	if err != nil {
		// L7: Log error for debugging
		slog.Error("blueprint_download_failed",
			"error", err.Error(),
			"document_url", documentURL,
		)
		return nil, fmt.Errorf("could not access blueprint: %w", err)
	}

	// Create multimodal request with image
	req := ai.NewMultimodalRequest(ai.ModelTypeFlash, prompt, imageData, mimeType)
	req.ReturnLogprobs = true

	result, err := s.aiClient.GenerateContent(ctx, req)
	if err != nil {
		// L7: Log AI failure
		slog.Error("ai_extraction_failed",
			"error", err.Error(),
			"document_url", documentURL,
		)
		return nil, fmt.Errorf("AI extraction failed: %w", err)
	}

	// Parse JSON response from Gemini
	// C2 Fix: Use square_footage (matches frontend CreateProjectRequest)
	var extraction struct {
		Name           string             `json:"name"`
		Address        string             `json:"address"`
		SquareFootage  float64            `json:"square_footage"`
		FoundationType string             `json:"foundation_type"`
		Stories        int                `json:"stories"`
		Bedrooms       int                `json:"bedrooms"`
		Bathrooms      int                `json:"bathrooms"`
		LongLeadItems  []struct {
			Name     string `json:"name"`
			Brand    string `json:"brand"`
			Model    string `json:"model"`
			Category string `json:"category"`
			Notes    string `json:"notes"`
		} `json:"long_lead_items"`
		Confidence map[string]float64 `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(result.Text), &extraction); err != nil {
		return nil, fmt.Errorf("failed to parse extraction result: %w", err)
	}

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

	// Convert extracted long-lead items with lead time estimates
	longLeadItems := s.enrichLongLeadItems(extraction.LongLeadItems)

	return &models.ExtractionResult{
		DocumentURL:   documentURL,
		ExtractedAt:   time.Now(),
		Values:        values,
		Confidence:    extraction.Confidence,
		LongLeadItems: longLeadItems,
	}, nil
}

// extractFromBytes uses Vision API to extract data from inline file bytes.
// Step 77: Direct file upload path that skips URL download.
func (s *InterrogatorService) extractFromBytes(
	ctx context.Context,
	fileData []byte,
	mimeType string,
	fileName string,
) (*models.ExtractionResult, error) {
	prompt := prompts.BlueprintExtractionPrompt()

	slog.Info("blueprint_extraction_from_bytes",
		"file_name", fileName,
		"mime_type", mimeType,
		"file_size", len(fileData),
	)

	// Create multimodal request with inline image
	req := ai.NewMultimodalRequest(ai.ModelTypeFlash, prompt, fileData, mimeType)
	req.ReturnLogprobs = true

	result, err := s.aiClient.GenerateContent(ctx, req)
	if err != nil {
		slog.Error("ai_extraction_failed_bytes",
			"error", err.Error(),
			"file_name", fileName,
		)
		return nil, fmt.Errorf("AI extraction failed: %w", err)
	}

	// Parse JSON response from Gemini (same structure as extractFromDocument)
	// C2 Fix: Use square_footage (matches frontend CreateProjectRequest)
	var extraction struct {
		Name           string             `json:"name"`
		Address        string             `json:"address"`
		SquareFootage  float64            `json:"square_footage"`
		FoundationType string             `json:"foundation_type"`
		Stories        int                `json:"stories"`
		Bedrooms       int                `json:"bedrooms"`
		Bathrooms      int                `json:"bathrooms"`
		LongLeadItems  []struct {
			Name     string `json:"name"`
			Brand    string `json:"brand"`
			Model    string `json:"model"`
			Category string `json:"category"`
			Notes    string `json:"notes"`
		} `json:"long_lead_items"`
		Confidence map[string]float64 `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(result.Text), &extraction); err != nil {
		return nil, fmt.Errorf("failed to parse extraction result: %w", err)
	}

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

	// Convert extracted long-lead items with lead time estimates
	longLeadItems := s.enrichLongLeadItems(extraction.LongLeadItems)

	return &models.ExtractionResult{
		DocumentURL:   fmt.Sprintf("inline:%s", fileName),
		ExtractedAt:   time.Now(),
		Values:        values,
		Confidence:    extraction.Confidence,
		LongLeadItems: longLeadItems,
	}, nil
}

// parseUserMessage extracts structured data from natural language.
func (s *InterrogatorService) parseUserMessage(
	ctx context.Context,
	message string,
	currentState map[string]interface{},
) (*models.ExtractionResult, error) {
	prompt := prompts.MessageParsingPrompt(message, currentState)

	req := ai.NewTextRequest(ai.ModelTypeFlash, prompt)
	req.ReturnLogprobs = true

	result, err := s.aiClient.GenerateContent(ctx, req)
	if err != nil {
		return nil, err
	}

	var extraction struct {
		Values     map[string]any     `json:"values"`
		Confidence map[string]float64 `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(result.Text), &extraction); err != nil {
		return nil, err
	}

	return &models.ExtractionResult{
		Values:     extraction.Values,
		Confidence: extraction.Confidence,
	}, nil
}

// getNextQuestion determines what to ask based on missing fields.
func (s *InterrogatorService) getNextQuestion(state map[string]interface{}) (string, string) {
	for _, field := range models.GetPriorityFields() {
		if _, exists := state[field.Field]; !exists {
			return field.Field, field.Question
		}
	}
	return "", ""
}

// checkReadyToCreate verifies minimum required fields.
func (s *InterrogatorService) checkReadyToCreate(state map[string]interface{}) bool {
	_, hasName := state["name"]
	_, hasAddress := state["address"]
	return hasName && hasAddress
}

// generateExtractionSummary creates a natural language summary of what was extracted.
func (s *InterrogatorService) generateExtractionSummary(extraction *models.ExtractionResult) string {
	var sb strings.Builder

	// Build summary of extracted fields
	sb.WriteString("I found these details from your plans:\n\n")

	if name, ok := extraction.Values["name"].(string); ok && name != "" {
		sb.WriteString(fmt.Sprintf("**Project**: %s\n", name))
	}
	if addr, ok := extraction.Values["address"].(string); ok && addr != "" {
		sb.WriteString(fmt.Sprintf("**Address**: %s\n", addr))
	}
	if sqft, ok := extraction.Values["square_footage"].(float64); ok && sqft > 0 {
		sb.WriteString(fmt.Sprintf("**Size**: %.0f sq ft\n", sqft))
	}
	if foundation, ok := extraction.Values["foundation_type"].(string); ok && foundation != "" {
		sb.WriteString(fmt.Sprintf("**Foundation**: %s\n", strings.Title(foundation)))
	}
	if stories, ok := extraction.Values["stories"].(int); ok && stories > 0 {
		sb.WriteString(fmt.Sprintf("**Stories**: %d\n", stories))
	}
	if bed, ok := extraction.Values["bedrooms"].(int); ok && bed > 0 {
		if bath, ok := extraction.Values["bathrooms"].(int); ok && bath > 0 {
			sb.WriteString(fmt.Sprintf("**%d bed / %d bath**\n", bed, bath))
		}
	}

	// Add long-lead item warnings
	if len(extraction.LongLeadItems) > 0 {
		sb.WriteString("\n**Long-lead items detected**:\n")
		for _, item := range extraction.LongLeadItems {
			if item.Brand != "" {
				sb.WriteString(fmt.Sprintf("- %s (%s) - ~%d weeks\n", item.Name, item.Brand, item.EstimatedLeadWeeks))
			} else {
				sb.WriteString(fmt.Sprintf("- %s - ~%d weeks\n", item.Name, item.EstimatedLeadWeeks))
			}
		}
	}

	sb.WriteString("\nWhen did your permit get issued, or when do you plan to break ground?")

	return sb.String()
}

// mergeStates combines current state with new extractions (new values win).
func (s *InterrogatorService) mergeStates(
	current map[string]interface{},
	extracted map[string]any,
) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range current {
		merged[k] = v
	}
	for k, v := range extracted {
		merged[k] = v
	}
	return merged
}

// downloadImage fetches an image from a URL with security controls.
// L7 Security: Prevents SSRF, enforces size limits, validates MIME types.
func (s *InterrogatorService) downloadImage(ctx context.Context, imageURL string) ([]byte, string, error) {
	// SECURITY: Validate URL scheme (block file://, ftp://, etc.)
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return nil, "", fmt.Errorf("invalid URL: %w", err)
	}

	// SECURITY: Only allow http/https (prevent file:// SSRF)
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, "", fmt.Errorf("unsupported URL scheme: %s (only http/https allowed)", parsedURL.Scheme)
	}

	// SECURITY: Block private IP ranges (prevent internal service access)
	// Pattern from vision_service.go:121-128
	host := parsedURL.Hostname()
	if isPrivateIP(host) {
		return nil, "", fmt.Errorf("access to private IP ranges forbidden")
	}

	// PERFORMANCE: Set 30-second timeout for download
	client := &http.Client{
		Timeout: 30 * time.Second,
		// SECURITY: Disable redirects to prevent redirect-based SSRF
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("image fetch failed with status: %d", resp.StatusCode)
	}

	// SECURITY: Validate MIME type before reading
	mimeType := resp.Header.Get("Content-Type")
	if !isValidImageMIME(mimeType) {
		return nil, "", fmt.Errorf("invalid MIME type: %s (expected image/jpeg, image/png, application/pdf)", mimeType)
	}

	// SECURITY: Enforce 50MB size limit (blueprints are typically 1-10MB)
	const maxBlueprintSize = 50 * 1024 * 1024 // 50MB
	limitedReader := io.LimitReader(resp.Body, maxBlueprintSize+1)

	imageData, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image: %w", err)
	}

	if len(imageData) > maxBlueprintSize {
		return nil, "", fmt.Errorf("image too large: %d bytes (max: %d)", len(imageData), maxBlueprintSize)
	}

	return imageData, mimeType, nil
}

// isPrivateIP checks if a hostname resolves to a private IP range.
// L7 Security: Prevents SSRF attacks against internal services.
func isPrivateIP(host string) bool {
	// Resolve hostname to IP
	ips, err := net.LookupIP(host)
	if err != nil {
		return true // Fail closed: if can't resolve, block it
	}

	for _, ip := range ips {
		// Check if IP is in private ranges
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
			return true
		}
		// Block AWS metadata endpoint (common SSRF target)
		if ip.String() == "169.254.169.254" {
			return true
		}
	}
	return false
}

// isValidImageMIME validates MIME type for blueprint uploads.
// C6 Fix: Uses exact match after stripping parameters to prevent bypass.
func isValidImageMIME(mimeType string) bool {
	// Strip parameters: "image/jpeg; charset=utf-8" → "image/jpeg"
	normalized := strings.TrimSpace(strings.SplitN(mimeType, ";", 2)[0])
	allowed := map[string]bool{
		"image/jpeg":      true,
		"image/jpg":       true,
		"image/png":       true,
		"image/webp":      true,
		"application/pdf": true,
	}
	return allowed[normalized]
}

// calculateAvgConfidence computes the average confidence score.
// L7: Used for observability metrics.
func calculateAvgConfidence(scores map[string]float64) float64 {
	if len(scores) == 0 {
		return 0.0
	}
	sum := 0.0
	for _, score := range scores {
		sum += score
	}
	return sum / float64(len(scores))
}

// enrichLongLeadItems converts raw extraction items to models with lead time estimates.
func (s *InterrogatorService) enrichLongLeadItems(items []struct {
	Name     string `json:"name"`
	Brand    string `json:"brand"`
	Model    string `json:"model"`
	Category string `json:"category"`
	Notes    string `json:"notes"`
}) []models.LongLeadItem {
	if len(items) == 0 {
		return nil
	}

	leadTimes := models.KnownBrandLeadTimes()
	result := make([]models.LongLeadItem, 0, len(items))

	for _, item := range items {
		leadWeeks := estimateLeadTime(item.Brand, item.Category, leadTimes)
		wbsCode := categoryToWBS(item.Category)

		result = append(result, models.LongLeadItem{
			Name:               item.Name,
			Brand:              item.Brand,
			Model:              item.Model,
			Category:           item.Category,
			EstimatedLeadWeeks: leadWeeks,
			WBSCode:            wbsCode,
			Notes:              item.Notes,
		})
	}

	return result
}

// estimateLeadTime determines lead time based on brand and category.
func estimateLeadTime(brand, category string, leadTimes map[string]int) int {
	// Normalize brand for lookup
	brandLower := strings.ToLower(strings.TrimSpace(brand))

	// Try exact brand match first
	if weeks, ok := leadTimes[brandLower]; ok {
		return weeks
	}

	// Try partial brand match
	for key, weeks := range leadTimes {
		if strings.Contains(brandLower, key) || strings.Contains(key, brandLower) {
			return weeks
		}
	}

	// Fall back to category defaults
	categoryDefaults := map[string]int{
		"windows":    8,
		"doors":      6,
		"hvac":       4,
		"appliances": 6,
		"millwork":   8,
		"finishes":   4,
	}

	if weeks, ok := categoryDefaults[strings.ToLower(category)]; ok {
		return weeks
	}

	return 4 // Default fallback
}

// categoryToWBS maps long-lead item categories to their typical WBS codes.
func categoryToWBS(category string) string {
	mapping := map[string]string{
		"windows":    "8.1", // Exterior Trim & Windows
		"doors":      "8.2", // Exterior Doors
		"hvac":       "9.1", // HVAC Rough-In
		"appliances": "14.1", // Appliance Installation
		"millwork":   "13.1", // Interior Trim & Doors
		"finishes":   "12.1", // Interior Paint
	}

	if wbs, ok := mapping[strings.ToLower(category)]; ok {
		return wbs
	}
	return ""
}
