package config

import (
	"errors"
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

	// SendGrid email configuration. See LAUNCH_STRATEGY.md A3.
	SendGridAPIKey   string
	EmailFromAddress string
	EmailFromName    string

	// BaseURL is the public URL of the application (e.g., https://app.futurebuild.app).
	// Used for constructing magic links and other URLs.
	BaseURL string

	// Twilio SMS configuration. See LAUNCH_PLAN.md P2 (Notifications/Toast UI).
	TwilioAccountSID  string
	TwilioAuthToken   string
	TwilioFromNumber  string

	// AuditWALPath is the file path for the audit Write-Ahead Log.
	// Used as a fallback when DB and DLQ are unavailable. See LAUNCH_PLAN.md.
	AuditWALPath string

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
		SendGridAPIKey:             os.Getenv("SENDGRID_API_KEY"),
		EmailFromAddress:           getEnvOrDefault("EMAIL_FROM_ADDRESS", "noreply@futurebuild.ai"),
		EmailFromName:              getEnvOrDefault("EMAIL_FROM_NAME", "FutureBuild"),
		BaseURL:                    getEnvOrDefault("BASE_URL", "http://localhost:8080"),
		TwilioAccountSID:           os.Getenv("TWILIO_ACCOUNT_SID"),
		TwilioAuthToken:            os.Getenv("TWILIO_AUTH_TOKEN"),
		TwilioFromNumber:           os.Getenv("TWILIO_FROM_NUMBER"),
		AuditWALPath:               getEnvOrDefault("AUDIT_WAL_PATH", "/var/log/futurebuild/audit.wal"),
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
	if c.JWTSecret == "" {
		return errors.New("JWT_SECRET environment variable is required")
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
