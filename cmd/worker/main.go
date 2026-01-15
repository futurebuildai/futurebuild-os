package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/colton/futurebuild/internal/agents"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/internal/worker"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Environment
	if err := godotenv.Load(); err != nil {
		log.Println("WARN: No .env file found, relying on environment variables")
	}

	// 2. Validate Required Configuration
	// See BACKEND_SCOPE.md Section 1 (Technology Stack)
	redisAddr := os.Getenv("REDIS_URL")
	dbURL := os.Getenv("DATABASE_URL")
	projectID := os.Getenv("GCP_PROJECT_ID")
	location := os.Getenv("GCP_LOCATION")

	if redisAddr == "" {
		log.Fatal("FATAL: REDIS_URL is required")
	}
	if dbURL == "" {
		log.Fatal("FATAL: DATABASE_URL is required")
	}
	if projectID == "" || location == "" {
		log.Println("WARN: GCP_PROJECT_ID or GCP_LOCATION not set. AI features will be disabled.")
	}

	// 3. Initialize Database
	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	// 4. Initialize AI Client (Optional - graceful degradation)
	// See BACKEND_SCOPE.md Section 3.2 (Context Engine)
	modelIDs := map[ai.ModelType]string{
		ai.ModelTypeFlash:     "gemini-2.0-flash-exp",
		ai.ModelTypePro:       "gemini-1.5-pro",
		ai.ModelTypeEmbedding: "text-embedding-004",
	}

	var aiClient ai.Client
	if projectID != "" && location != "" {
		aiClient, err = ai.NewVertexClient(context.Background(), projectID, location, modelIDs)
		if err != nil {
			log.Printf("WARN: Failed to create Vertex Client: %v (AI features will fail)", err)
		}
	}

	// 5. Initialize Services
	scheduleService := service.NewScheduleService(dbPool)
	notificationService := service.NewConsoleEmailProvider()
	weatherService := service.NewMockWeatherService()

	// 6. Initialize Agents
	dailyFocusAgent := agents.NewDailyFocusAgent(
		dbPool,
		scheduleService,
		weatherService,
		notificationService,
		aiClient,
	)

	// 7. Initialize Worker Handlers
	workerHandler := worker.NewWorkerHandler(dailyFocusAgent)

	// 8. Initialize Scheduler (The Clock)
	// See PRODUCTION_PLAN.md Step 45 (Daily Briefing Job)
	scheduler := worker.NewScheduler(redisAddr)

	if _, err := scheduler.RegisterEntry("0 6 * * *", worker.NewDailyBriefingTask()); err != nil {
		log.Fatalf("could not register daily briefing cron: %v", err)
	}

	// 9. Initialize Worker Server (The Processor)
	srv := worker.NewServer(redisAddr, 10)

	// Register the actual handler function
	srv.RegisterHandlerFunc(worker.TypeDailyBriefing, workerHandler.HandleDailyBriefing)

	// 10. Start Services with Error Propagation
	// Both scheduler and server run in goroutines.
	// If either fails, we propagate the error to main for graceful shutdown.
	errChan := make(chan error, 2)

	go func() {
		if err := scheduler.Run(); err != nil {
			errChan <- fmt.Errorf("scheduler: %w", err)
		}
	}()

	go func() {
		if err := srv.Run(); err != nil {
			errChan <- fmt.Errorf("worker server: %w", err)
		}
	}()

	log.Println("FutureBuild Worker Started. Press Ctrl+C to stop.")

	// 11. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Wait for either shutdown signal OR critical error
	select {
	case err := <-errChan:
		log.Fatalf("FATAL: %v", err)
	case <-quit:
		log.Println("Shutting down worker...")
	}
	scheduler.Shutdown()
	srv.Shutdown()
	log.Println("Worker stopped")
}
