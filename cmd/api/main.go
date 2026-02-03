package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/platform/errormon"
	"github.com/colton/futurebuild/internal/server"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Initialize error monitoring (slog-based default; swap for Sentry/Datadog in production).
	// See L7 Gate Audit Item 15.
	errormon.Init(errormon.NewSlogReporter(nil))
	defer errormon.Get().Flush()

	ctx := context.Background()

	// Initialize Vertex AI Client
	modelIDs := map[ai.ModelType]string{
		ai.ModelTypeFlash:     cfg.VertexModelFlashID,
		ai.ModelTypePro:       cfg.VertexModelProID,
		ai.ModelTypeEmbedding: cfg.VertexModelEmbeddingID,
	}
	aiClient, err := ai.NewVertexClient(ctx, cfg.VertexProjectID, cfg.VertexLocation, modelIDs)
	if err != nil {
		log.Fatalf("Failed to initialize Vertex AI client: %v", err)
	}
	defer aiClient.Close()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Database connection check failed: %v", err)
	}

	fmt.Println("Database connection established")

	srv := server.NewServer(pool, cfg, aiClient)

	// Create HTTP server with production-safe timeouts.
	// See Staging Readiness Audit: graceful shutdown + request timeouts.
	httpServer := srv.NewHTTPServer()

	// Start server in goroutine so we can handle shutdown signals.
	go func() {
		fmt.Printf("Server starting on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for SIGTERM or SIGINT for graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server stopped")
}
