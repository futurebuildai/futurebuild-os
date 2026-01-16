// Package ai provides vendor-agnostic AI client abstractions.
// This package enables switching between AI providers (Google Vertex AI, OpenAI,
// Anthropic, local models) without modifying service layer code.
//
// L7 Quality Gate: Vendor abstraction eliminates genai imports outside this package.
package ai

// =============================================================================
// VENDOR-AGNOSTIC TYPES
// =============================================================================
// These types abstract away vendor-specific representations (e.g., *genai.Part)
// from the service layer, enabling clean dependency injection and testing.
// =============================================================================

// ContentPart represents a single part of multimodal AI input.
// Can contain text, image data, or other media.
type ContentPart struct {
	// Text is the text content of this part.
	// Mutually exclusive with Data.
	Text string

	// MimeType specifies the media type when Data is present.
	// Examples: "image/jpeg", "image/png", "application/pdf"
	MimeType string

	// Data contains binary data for images, documents, etc.
	// Mutually exclusive with Text.
	Data []byte
}

// GenerateRequest encapsulates all parameters for content generation.
type GenerateRequest struct {
	// Model specifies which AI model to use.
	Model ModelType

	// Parts contains the multimodal input (text, images, etc.)
	Parts []ContentPart

	// MaxTokens limits the response length.
	// Zero means use model default.
	MaxTokens int

	// Temperature controls randomness (0.0 = deterministic, 1.0 = creative).
	// Zero means use model default.
	Temperature float32
}

// GenerateResponse contains the result of content generation.
type GenerateResponse struct {
	// Text is the generated text content.
	Text string

	// TokensUsed is the total token count (input + output).
	TokensUsed int

	// Confidence is an optional model confidence score (0.0 to 1.0).
	Confidence float32
}

// NewTextRequest creates a simple text-only GenerateRequest.
// Convenience function for common use cases.
func NewTextRequest(model ModelType, text string) GenerateRequest {
	return GenerateRequest{
		Model: model,
		Parts: []ContentPart{{Text: text}},
	}
}

// NewMultimodalRequest creates a GenerateRequest with text and image.
// Convenience function for vision use cases.
func NewMultimodalRequest(model ModelType, text string, imageData []byte, mimeType string) GenerateRequest {
	return GenerateRequest{
		Model: model,
		Parts: []ContentPart{
			{Text: text},
			{Data: imageData, MimeType: mimeType},
		},
	}
}
