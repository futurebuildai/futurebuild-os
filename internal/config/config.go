package config

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration.
// See PRODUCTION_PLAN.md Task 3 (Harden Configuration).
type Config struct {
	DatabaseURL            string
	RedisURL               string
	AppPort                int
	JWTSecret              string
	JWTExpiry              time.Duration
	TrustedProxies         []string
	VertexProjectID        string
	VertexLocation         string
	VertexModelFlashID     string
	VertexModelProID       string
	VertexModelEmbeddingID string
	S3Endpoint             string
	S3Bucket               string
	S3AccessKey            string
	S3SecretKey            string
	WebhookSecret          string // See PRODUCTION_PLAN.md Step 48 (Signature Verification)
	// Environment specifies the runtime environment (development, staging, production).
	// Used for safety checks like NoOp implementations. See Code Review Issue 3B.
	Environment string

	// InvoiceConfidenceThreshold defines the minimum AI confidence required to bypass human review.
	// Defaults to 0.85. See Code Review Issue 3B.
	InvoiceConfidenceThreshold float64

	// FutureShade configuration. See FUTURESHADE_INIT_specs.md.
	// FutureShadeEnabled toggles the FutureShade intelligence layer.
	FutureShadeEnabled bool
	// FutureShadeAPIKey is the API key for the AI provider used by FutureShade.
	FutureShadeAPIKey string
	// FutureShadeModelID is the model ID for FutureShade analysis.
	FutureShadeModelID string

	// ProjectRoot is the root directory of the project.
	// Used by ShadowDocs to serve documentation files. See SHADOW_VIEWER_specs.md.
	ProjectRoot string

	// Email provider configuration. See LAUNCH_STRATEGY.md A3.
	ResendAPIKey     string
	EmailFromAddress string
	EmailFromName    string

	// BaseURL is the public URL of the application (e.g., https://app.futurebuild.app).
	// Used for constructing portal magic links and other URLs.
	BaseURL string

	// Twilio SMS configuration. See LAUNCH_PLAN.md P2 (Notifications/Toast UI).
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromNumber string

	// Bird (MessageBird) configuration. Unified SMS + Email provider.
	// See migration plan for Twilio/Resend → Bird.
	BirdAccessKey        string
	BirdOriginator       string // SMS sender ID (alphanumeric or phone number)
	NotificationProvider string // "bird", "legacy", or "console"

	// AuditWALPath is the file path for the audit Write-Ahead Log.
	// Used as a fallback when DB and DLQ are unavailable. See LAUNCH_PLAN.md.
	AuditWALPath string

	// Clerk Identity Provider configuration. See STEP_78_AUTH_PROVIDER.md.
	// JWKS URL is derived as {ClerkIssuerURL}/.well-known/jwks.json
	ClerkIssuerURL string
	// ClerkAudience is the expected audience (aud) claim for Clerk JWTs.
	// Optional: only enforced when non-empty. Configure in Clerk Dashboard JWT template.
	// See STEP_79_MIDDLEWARE_SWAP.md Section 1.2.
	ClerkAudience string
	// ClerkWebhookSecret is the Svix signing secret for Clerk webhook verification.
	// Optional: webhook handler is fail-closed when empty (rejects all requests).
	// See PHASE_12_PRD.md Step 80: Organization Manager.
	ClerkWebhookSecret string
	// ClerkSecretKey is the Clerk Backend API secret key (sk_xxx).
	// Used to create users and manage org memberships during invite acceptance.
	ClerkSecretKey string

	// GitHub configuration for Automated PR Review.
	// See docs/AUTOMATED_PR_REVIEW_PRD.md
	GitHubWebhookSecret string // GITHUB_WEBHOOK_SECRET - HMAC verification
	GitHubPAT           string // GITHUB_PAT - API authentication

	// Anthropic/Claude configuration for agentic layer.
	// Claude Opus 4.6 sits alongside Gemini — handles reasoning about
	// critical path, budget, onboarding, and chat intelligence.
	AnthropicAPIKey string // ANTHROPIC_API_KEY
	ClaudeModelID   string // CLAUDE_MODEL_ID (default: claude-opus-4-6-20250918)

	// FB-Brain connection for cross-system integration.
	BrainURL            string // FB_BRAIN_URL
	BrainIntegrationKey string // FB_BRAIN_INTEGRATION_KEY

	Worker WorkerConfig
}

type WorkerConfig struct {
	QueuePriorities map[string]int
	Concurrency     int
}

// LoadConfig loads configuration from environment variables.
// Returns an error if critical configuration is missing (Fail Fast).
// See PRODUCTION_PLAN.md Task 3.
func LoadConfig() (*Config, error) {
	portStr := os.Getenv("APP_PORT")
	if portStr == "" {
		portStr = os.Getenv("PORT") // Railway injects PORT
	}
	if portStr == "" {
		portStr = "8080"
	}
	port, _ := strconv.Atoi(portStr)

	expiryStr := os.Getenv("JWT_EXPIRY")
	if expiryStr == "" {
		expiryStr = "24h"
	}
	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		expiry = 24 * time.Hour
	}

	// Parse InvoiceConfidenceThreshold with default 0.85
	confidenceStr := getEnvOrDefault("INVOICE_CONFIDENCE_THRESHOLD", "0.85")
	confidenceThreshold, err := strconv.ParseFloat(confidenceStr, 64)
	if err != nil {
		confidenceThreshold = 0.85
	}

	// Handle Service Account JSON from environment variable if provided.
	// This allows deployment to platforms like DigitalOcean App Platform without file mounting.
	if saContent := os.Getenv("GCP_SA_JSON_CONTENT"); saContent != "" {
		log.Printf("GCP_SA_JSON_CONTENT found (%d bytes), writing to temp file", len(saContent))
		saPath := "/tmp/service-account.json"
		if err := os.WriteFile(saPath, []byte(saContent), 0600); err != nil {
			log.Printf("ERROR: Failed to write service account JSON to %s: %v", saPath, err)
		} else {
			os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", saPath)
			log.Printf("GOOGLE_APPLICATION_CREDENTIALS set to %s", saPath)
		}
	} else {
		log.Println("WARNING: GCP_SA_JSON_CONTENT not set - Vertex AI may fail to authenticate")
	}

	cfg := &Config{
		DatabaseURL:                os.Getenv("DATABASE_URL"),
		RedisURL:                   getEnvOrDefault("REDIS_URL", "localhost:6379"),
		AppPort:                    port,
		JWTSecret:                  os.Getenv("JWT_SECRET"),
		JWTExpiry:                  expiry,
		TrustedProxies:             strings.Split(os.Getenv("TRUSTED_PROXIES"), ","),
		VertexProjectID:            os.Getenv("VERTEX_PROJECT_ID"),
		VertexLocation:             os.Getenv("VERTEX_LOCATION"),
		VertexModelFlashID:         getEnvOrDefault("VERTEX_MODEL_FLASH_ID", "gemini-2.5-flash"),
		VertexModelProID:           getEnvOrDefault("VERTEX_MODEL_PRO_ID", "gemini-1.5-pro"),
		VertexModelEmbeddingID:     getEnvOrDefault("VERTEX_MODEL_EMBEDDING_ID", "text-embedding-004"),
		S3Endpoint:                 os.Getenv("S3_ENDPOINT"),
		S3Bucket:                   os.Getenv("S3_BUCKET"),
		S3AccessKey:                os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey:                os.Getenv("S3_SECRET_KEY"),
		WebhookSecret:              os.Getenv("WEBHOOK_SECRET"),
		Environment:                getEnvOrDefault("APP_ENV", "development"),
		InvoiceConfidenceThreshold: confidenceThreshold,
		FutureShadeEnabled:         getEnvBool("FUTURESHADE_ENABLED", false),
		FutureShadeAPIKey:          os.Getenv("FUTURESHADE_API_KEY"),
		FutureShadeModelID:         getEnvOrDefault("FUTURESHADE_MODEL_ID", "gemini-2.5-flash"),
		ProjectRoot:                getEnvOrDefault("PROJECT_ROOT", "."),
		ResendAPIKey:               os.Getenv("RESEND_API_KEY"),
		EmailFromAddress:           getEnvOrDefault("EMAIL_FROM_ADDRESS", "noreply@futurebuild.ai"),
		EmailFromName:              getEnvOrDefault("EMAIL_FROM_NAME", "FutureBuild"),
		BaseURL:                    getEnvOrDefault("BASE_URL", "http://localhost:8080"),
		TwilioAccountSID:           os.Getenv("TWILIO_ACCOUNT_SID"),
		TwilioAuthToken:            os.Getenv("TWILIO_AUTH_TOKEN"),
		TwilioFromNumber:           os.Getenv("TWILIO_FROM_NUMBER"),
		BirdAccessKey:              os.Getenv("BIRD_ACCESS_KEY"),
		BirdOriginator:             getEnvOrDefault("BIRD_ORIGINATOR", "FutureBuild"),
		NotificationProvider:       getEnvOrDefault("NOTIFICATION_PROVIDER", "legacy"),
		AuditWALPath:               getEnvOrDefault("AUDIT_WAL_PATH", "/var/lib/futurebuild/audit.wal"),
		ClerkIssuerURL:             os.Getenv("CLERK_ISSUER_URL"),
		ClerkAudience:              os.Getenv("CLERK_AUDIENCE"),
		ClerkWebhookSecret:         os.Getenv("CLERK_WEBHOOK_SECRET"),
		ClerkSecretKey:             os.Getenv("CLERK_SECRET_KEY"),
		GitHubWebhookSecret:        os.Getenv("GITHUB_WEBHOOK_SECRET"),
		GitHubPAT:                  os.Getenv("GITHUB_PAT"),
		AnthropicAPIKey:            os.Getenv("ANTHROPIC_API_KEY"),
		ClaudeModelID:              getEnvOrDefault("CLAUDE_MODEL_ID", "claude-opus-4-6-20250918"),
		BrainURL:                   os.Getenv("FB_BRAIN_URL"),
		BrainIntegrationKey:        os.Getenv("FB_BRAIN_INTEGRATION_KEY"),
		Worker: WorkerConfig{
			Concurrency: 10,
			QueuePriorities: map[string]int{
				"critical": getEnvInt("WORKER_QUEUE_CRITICAL", 6),
				"default":  getEnvInt("WORKER_QUEUE_DEFAULT", 3),
				"low":      getEnvInt("WORKER_QUEUE_LOW", 1),
			},
		},
	}

	// Fail Fast: Validate critical configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that all required configuration is present.
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL environment variable is required")
	}
	// Phase 12: JWT_SECRET is no longer required — Clerk uses JWKS (RS256).
	// CLERK_ISSUER_URL is required for JWKS-based JWT validation,
	// unless DEV_AUTH_BYPASS is enabled (demo/development mode).
	if c.ClerkIssuerURL == "" && os.Getenv("DEV_AUTH_BYPASS") != "true" {
		return errors.New("CLERK_ISSUER_URL environment variable is required")
	}
	return nil
}

func getEnvOrDefault(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if valStr := os.Getenv(key); valStr != "" {
		if val, err := strconv.Atoi(valStr); err == nil {
			return val
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if valStr := os.Getenv(key); valStr != "" {
		val, err := strconv.ParseBool(valStr)
		if err == nil {
			return val
		}
	}
	return fallback
}

// IsProduction returns true if the environment is production.
func (c *Config) IsProduction() bool {
	return c.Environment == "production" || c.Environment == "prod"
}

// IsStaging returns true if the environment is staging.
func (c *Config) IsStaging() bool {
	return c.Environment == "staging" || c.Environment == "stage"
}

// IsDemo returns true if the environment is demo.
func (c *Config) IsDemo() bool {
	return c.Environment == "demo"
}

// IsDevelopment returns true if the environment is development.
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development" || c.Environment == "dev"
}
