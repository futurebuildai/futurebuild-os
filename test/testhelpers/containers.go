// Package testhelpers provides shared testing utilities.
// See PRODUCTION_PLAN.md: Testing Strategy & CI Reliability Remediation
package testhelpers

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// StartRedisContainer starts an ephemeral Redis container for Asynq.
// Returns the connection string (host:port) and a cleanup function.
func StartRedisContainer(t *testing.T) (string, func()) {
	t.Helper()

	if os.Getenv("SKIP_TESTCONTAINERS") == "1" {
		addr := os.Getenv("REDIS_ADDR")
		if addr == "" {
			addr = "localhost:6379"
		}
		return addr, func() {}
	}

	ctx := context.Background()
	container, err := redis.Run(ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections"),
		),
	)
	if err != nil {
		t.Fatalf("Failed to start Redis container: %v", err)
	}

	endpoint, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("Failed to get Redis connection string: %v", err)
	}

	// Strip scheme for go-redis Addr compatibility
	endpoint = strings.TrimPrefix(endpoint, "redis://")

	return endpoint, func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate Redis container: %v", err)
		}
	}
}

// StartPostgresContainer starts an ephemeral PostgreSQL container for integration tests.
// Automatically applies all migrations from db/migrations/.
// Returns the connection pool and a cleanup function.
//
// Environment:
//   - SKIP_TESTCONTAINERS=1: Use DATABASE_URL environment variable instead of container
//
// FAANG Standard: Tests should run in a clean, isolated state.
func StartPostgresContainer(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	// Allow skipping testcontainers for local development with existing DB
	if os.Getenv("SKIP_TESTCONTAINERS") == "1" {
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			dbURL = "postgres://fb_user:fb_pass@localhost:5433/futurebuild?sslmode=disable"
		}
		ctx := context.Background()
		pool, err := pgxpool.New(ctx, dbURL)
		if err != nil {
			t.Fatalf("Failed to connect to external database: %v", err)
		}
		return pool, func() { pool.Close() }
	}

	ctx := context.Background()

	// Find migrations directory (relative to repo root)
	migrationsDir := findMigrationsDir(t)

	// Start PostgreSQL container with pgvector extension and migrations
	// Use pgvector/pgvector image which includes the vector extension
	container, err := postgres.Run(ctx,
		"pgvector/pgvector:pg16",
		postgres.WithDatabase("futurebuild_test"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_pass"),
		postgres.WithInitScripts(getMigrationScripts(t, migrationsDir)...),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	// Get connection string
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Create connection pool
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to create connection pool: %v", err)
	}

	// L7 Hardening: Validate pgvector extension is available
	// Panic early if the wrong Postgres image is used (without vector support)
	var extExists bool
	err = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'vector')").Scan(&extExists)
	if err != nil {
		t.Fatalf("Failed to check pgvector extension: %v", err)
	}
	if !extExists {
		t.Fatalf("pgvector extension missing in test container")
	}

	// Return cleanup function
	cleanup := func() {
		pool.Close()
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Warning: Failed to terminate container: %v", err)
		}
	}

	return pool, cleanup
}

// findMigrationsDir locates the db/migrations directory from test context.
func findMigrationsDir(t *testing.T) string {
	t.Helper()

	// Try common relative paths from test directories
	candidates := []string{
		"../../migrations",       // From test/simulation/
		"../../../migrations",    // From nested test dirs
		"migrations",             // From repo root
		"../../db/migrations",    // Legacy: from test/simulation/
		"../../../db/migrations", // Legacy: from nested dirs
		"db/migrations",          // Legacy: from repo root
	}

	// Also try using go module root
	cwd, _ := os.Getwd()
	for _, candidate := range candidates {
		absPath := filepath.Join(cwd, candidate)
		if info, err := os.Stat(absPath); err == nil && info.IsDir() {
			return absPath
		}
	}

	t.Fatalf("Could not find db/migrations directory from %s", cwd)
	return ""
}

// getMigrationScripts returns all .sql migration files in order.
func getMigrationScripts(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Failed to read migrations directory: %v", err)
	}

	var scripts []string
	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			// Only include "up" migrations (e.g., 000001_init.up.sql)
			// Exclude "down" migrations
			if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
				if filepath.Ext(name) == ".sql" && len(name) > 7 && name[len(name)-7:] == ".up.sql" {
					scripts = append(scripts, filepath.Join(dir, name))
				}
			}
		}
	}

	return scripts
}
