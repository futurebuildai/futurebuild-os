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
		"current_fields", len(req.CurrentState),
	)

	resp := &models.OnboardResponse{
		SessionID:        req.SessionID,
		ExtractedValues:  make(map[string]any),
		ConfidenceScores: make(map[string]float64),
		ReadyToCreate:    false,
	}

	// BRANCH 1: Document uploaded → Extract via Vision API
	if req.DocumentURL != "" {
		extraction, err := s.extractFromDocument(ctx, req.DocumentURL)
		if err != nil {
			resp.Reply = "I couldn't read that file. Could you try a clearer scan or describe the project?"
			return resp, nil
		}

		// Merge extracted values into response
		for k, v := range extraction.Values {
			resp.ExtractedValues[k] = v
			resp.ConfidenceScores[k] = extraction.Confidence[k]
		}

		// Generate a summary message
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
	var extraction struct {
		Name           string             `json:"name"`
		Address        string             `json:"address"`
		GSF            float64            `json:"gsf"`
		FoundationType string             `json:"foundation_type"`
		Stories        int                `json:"stories"`
		Bedrooms       int                `json:"bedrooms"`
		Bathrooms      int                `json:"bathrooms"`
		Confidence     map[string]float64 `json:"confidence"`
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
	if extraction.GSF > 0 {
		values["gsf"] = extraction.GSF
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

	return &models.ExtractionResult{
		DocumentURL: documentURL,
		ExtractedAt: time.Now(), // L7: Add timestamp
		Values:      values,
		Confidence:  extraction.Confidence,
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
	count := len(extraction.Values)
	return fmt.Sprintf("I found %d details from your blueprint. Review them in the form and let me know if anything needs to be corrected.", count)
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
func isValidImageMIME(mimeType string) bool {
	allowed := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/webp",
		"application/pdf",
	}
	for _, allowedType := range allowed {
		if strings.HasPrefix(mimeType, allowedType) {
			return true
		}
	}
	return false
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
