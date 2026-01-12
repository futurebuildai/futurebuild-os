package main

import (
	"context"
	"fmt"
	"log"

	"github.com/colton/futurebuild/internal/config"
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

	cfg := config.LoadConfig()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

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
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
