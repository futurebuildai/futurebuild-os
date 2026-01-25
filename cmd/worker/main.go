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
	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/internal/worker"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/colton/futurebuild/pkg/clock"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Configuration (Centralized)
	if err := godotenv.Load(); err != nil {
		log.Println("WARN: No .env file found, relying on environment variables")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("FATAL: Failed to load configuration: %v", err)
	}

	// 2. Setup Dependencies
	// See BACKEND_SCOPE.md Section 1 (Technology Stack)
	// AI Features Check
	if cfg.VertexProjectID == "" || cfg.VertexLocation == "" {
		log.Println("WARN: GCP_PROJECT_ID or GCP_LOCATION not set. AI features will be disabled.")
	}

	// 3. Initialize Database
	dbPool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	// 4. Initialize AI Client (Optional - graceful degradation)
	// See BACKEND_SCOPE.md Section 3.2 (Context Engine)
	modelIDs := map[ai.ModelType]string{
		ai.ModelTypeFlash:     cfg.VertexModelFlashID,
		ai.ModelTypePro:       cfg.VertexModelProID,
		ai.ModelTypeEmbedding: cfg.VertexModelEmbeddingID,
	}

	var aiClient ai.Client
	if cfg.VertexProjectID != "" && cfg.VertexLocation != "" {
		aiClient, err = ai.NewVertexClient(context.Background(), cfg.VertexProjectID, cfg.VertexLocation, modelIDs)
		if err != nil {
			log.Printf("WARN: Failed to create Vertex Client: %v (AI features will fail)", err)
		}
	}

	// 5. Initialize Services
	projectService := service.NewProjectService(dbPool)
	scheduleService := service.NewScheduleService(dbPool)
	notificationService := service.NewConsoleEmailProvider()
	weatherService := service.NewMockWeatherService()
	// Critical Blocker A Remediation: Add geocoding and directory services
	geocodingService := service.NewGeocodingService()
	directoryService := service.NewDirectoryService(dbPool)

	// 6. Initialize Agents
	// See PRODUCTION_PLAN.md Step 49: Using RealClock for production
	realClock := clock.RealClock{}

	dailyFocusAgent := agents.NewDailyFocusAgentWithService(
		projectService, // Replaces dbPool - Clean Service Layer Pattern
		scheduleService,
		weatherService,
		notificationService,
		aiClient,
		realClock,
		geocodingService, // Critical Blocker A
		directoryService, // Critical Blocker A
	)

	// Procurement Agent for long-lead item monitoring
	// See PRODUCTION_PLAN.md Step 46, 49
	// Config Decoupling: Load ProcurementConfig from environment (defaults if not set)
	procurementCfg := config.LoadProcurementConfigFromEnv()
	procurementAgent := agents.NewProcurementAgentWithDB(dbPool, weatherService, realClock, procurementCfg)

	// 7. Initialize Worker Handlers
	// P1 Performance Fix: Pass db and clock for notification handler
	workerHandler := worker.NewWorkerHandler(dailyFocusAgent, procurementAgent, dbPool, realClock)

	// 8. Initialize Scheduler (The Clock)
	// See PRODUCTION_PLAN.md Step 45 (Daily Briefing Job)
	scheduler := worker.NewScheduler(cfg.RedisURL)

	if _, err := scheduler.RegisterEntry("0 6 * * *", worker.NewDailyBriefingTask()); err != nil {
		log.Fatalf("could not register daily briefing cron: %v", err)
	}

	// Register Procurement Check at 05:00 AM UTC (before daily briefing)
	// See PRODUCTION_PLAN.md Step 46
	if _, err := scheduler.RegisterEntry("0 5 * * *", worker.NewProcurementCheckTask()); err != nil {
		log.Fatalf("could not register procurement check cron: %v", err)
	}

	// 9. Initialize Worker Server (The Processor)
	// L7 Config: Use configured priorities. Default to 10 concurrency if not set (though config loader sets default)
	concurrency := 10
	if cfg.Worker.Concurrency > 0 {
		concurrency = cfg.Worker.Concurrency
	}
	// Fallback to defaults if map is nil/empty (safety)
	queues := cfg.Worker.QueuePriorities
	if len(queues) == 0 {
		queues = map[string]int{"critical": 6, "default": 3, "low": 1}
	}
	srv := worker.NewServer(cfg.RedisURL, concurrency, queues)

	// Register the actual handler function
	srv.RegisterHandlerFunc(worker.TypeDailyBriefing, workerHandler.HandleDailyBriefing)
	srv.RegisterHandlerFunc(worker.TypeProcurementCheck, workerHandler.HandleProcurementCheck)
	srv.RegisterHandlerFunc(worker.TypeHydrateProject, workerHandler.HandleHydrateProject)
	// P1 Performance Fix: Register async notification handler (sidecar pattern)
	srv.RegisterHandlerFunc(worker.TypeProcurementNotification, workerHandler.HandleProcurementNotification)

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
