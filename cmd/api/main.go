package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/platform/errormon"
	"github.com/colton/futurebuild/internal/readiness"
	"github.com/colton/futurebuild/internal/server"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	readinessCheck := flag.Bool("readiness-check", false, "Run integration readiness checks and exit")
	flag.Parse()

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Handle --readiness-check: build probes from config, run checks, print JSON, exit.
	if *readinessCheck {
		runReadinessCheckAndExit(cfg)
		return
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

// runReadinessCheckAndExit builds a readiness service from config, runs all probes,
// prints the JSON report to stdout, and exits with 0 (healthy/degraded) or 1 (failed).
// This is used by CI/CD pipelines and entrypoint.sh for pre-flight checks.
func runReadinessCheckAndExit(cfg *config.Config) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a DB pool just for the probe (matches server config path).
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("readiness-check: failed to create DB pool: %v", err)
	}
	defer pool.Close()

	svc := readiness.NewService(15*time.Second,
		readiness.NewDatabaseProbe(pool),
		readiness.NewClerkProbe(cfg.ClerkIssuerURL),
		readiness.NewRedisProbe(cfg.RedisURL),
		readiness.NewResendProbe(cfg.ResendAPIKey),
		readiness.NewSendGridProbe(cfg.SendGridAPIKey),
		readiness.NewTwilioProbe(cfg.TwilioAccountSID, cfg.TwilioAuthToken),
		readiness.NewVertexAIProbe(cfg.VertexProjectID, cfg.VertexLocation),
		readiness.NewS3Probe(cfg.S3Endpoint, cfg.S3Bucket, cfg.S3AccessKey, cfg.S3SecretKey),
	)

	report := svc.Run(ctx, cfg.Environment)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(report)

	if report.Status == readiness.StatusFailed {
		os.Exit(1)
	}
}
