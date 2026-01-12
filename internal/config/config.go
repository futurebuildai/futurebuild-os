package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL            string
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
}

func LoadConfig() *Config {
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

	return &Config{
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		AppPort:                port,
		JWTSecret:              os.Getenv("JWT_SECRET"),
		JWTExpiry:              expiry,
		TrustedProxies:         strings.Split(os.Getenv("TRUSTED_PROXIES"), ","),
		VertexProjectID:        os.Getenv("VERTEX_PROJECT_ID"),
		VertexLocation:         os.Getenv("VERTEX_LOCATION"),
		VertexModelFlashID:     getEnvOrDefault("VERTEX_MODEL_FLASH_ID", "gemini-2.5-flash"),
		VertexModelProID:       getEnvOrDefault("VERTEX_MODEL_PRO_ID", "gemini-1.5-pro"),
		VertexModelEmbeddingID: getEnvOrDefault("VERTEX_MODEL_EMBEDDING_ID", "text-embedding-004"),
		S3Endpoint:             os.Getenv("S3_ENDPOINT"),
		S3Bucket:               os.Getenv("S3_BUCKET"),
		S3AccessKey:            os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey:            os.Getenv("S3_SECRET_KEY"),
	}
}

func getEnvOrDefault(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
