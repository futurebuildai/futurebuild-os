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
