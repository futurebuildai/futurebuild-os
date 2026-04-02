package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all environment-based configuration for FB-Brain.
type Config struct {
	// Database
	DatabaseURL string
	DBPoolMax   int32
	DBPoolMin   int32
	DBTimeout   time.Duration

	// Server
	Port   string
	Issuer string // OIDC issuer URL (e.g. https://brain.futurebuild.io)

	// OIDC
	CryptoKeyHex string // 32-byte hex key for OIDC token encryption
	DevMode      bool   // Allow http:// redirect URIs and insecure issuer

	// Redis (future use)
	RedisURL string
}

// Load parses environment variables into a Config struct.
func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	issuer := os.Getenv("ISSUER_URL")
	if issuer == "" {
		issuer = "http://localhost:8081"
	}

	cfg := &Config{
		DatabaseURL:  dbURL,
		DBPoolMax:    envInt32("DB_POOL_MAX", 25),
		DBPoolMin:    envInt32("DB_POOL_MIN", 5),
		DBTimeout:    envDuration("DB_TIMEOUT", 5*time.Second),
		Port:         envString("PORT", "8081"),
		Issuer:       issuer,
		CryptoKeyHex: envString("OIDC_CRYPTO_KEY", ""),
		DevMode:      envBool("DEV_MODE", false),
		RedisURL:     envString("REDIS_URL", ""),
	}

	return cfg, nil
}

func envString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt32(key string, fallback int32) int32 {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return int32(n)
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return fallback
}
