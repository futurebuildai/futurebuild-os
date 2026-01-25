package futureshade

// Config holds configuration for the FutureShade service.
type Config struct {
	// APIKey is the API key for the AI provider.
	APIKey string
	// ModelID is the model ID to use for analysis.
	ModelID string
	// Enabled determines if the service is active.
	Enabled bool
}
