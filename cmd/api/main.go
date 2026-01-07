package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/colton/futurebuild/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	portStr := os.Getenv("APP_PORT")
	if portStr == "" {
		portStr = "8080" // Default port
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid APP_PORT: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Database connection check failed: %v", err)
	}

	fmt.Println("Database connection established")

	srv := server.NewServer(pool, port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
