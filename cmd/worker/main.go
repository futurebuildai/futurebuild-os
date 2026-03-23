package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/colton/futurebuild/internal/adapters"
	"github.com/colton/futurebuild/internal/agents"
	"github.com/colton/futurebuild/internal/agents/tools"
	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/futureshade"
	"github.com/colton/futurebuild/internal/futureshade/gateway"
	"github.com/colton/futurebuild/internal/futureshade/skills"
	"github.com/colton/futurebuild/internal/futureshade/tribunal"
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
	notificationService := service.NewNotificationService(
		cfg.ResendAPIKey,
		cfg.EmailFromAddress,
		cfg.EmailFromName,
		cfg.TwilioAccountSID,
		cfg.TwilioAuthToken,
		cfg.TwilioFromNumber,
		cfg.BirdAccessKey,
		cfg.BirdOriginator,
		cfg.NotificationProvider,
	)
	weatherService := service.NewMockWeatherService()
	// Critical Blocker A Remediation: Add geocoding and directory services
	geocodingService := service.NewGeocodingService()
	directoryService := service.NewDirectoryService(dbPool)
	// V2 Feed: FeedService for agent card population
	feedService := service.NewFeedService(dbPool)

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
	).WithFeedWriter(feedService) // V2 Feed: Write daily_briefing cards

	// Agent action service for human-in-the-loop approval (worker needs this for approval cards)
	agentActionService := service.NewAgentActionService(dbPool, feedService)

	// Claude-powered reasoning layer (optional — graceful degradation to Gemini)
	if cfg.AnthropicAPIKey != "" {
		claudeModelMap := map[ai.ModelType]string{
			ai.ModelTypeOpus: cfg.ClaudeModelID,
		}
		claudeClient := ai.NewAnthropicClient(cfg.AnthropicAPIKey, claudeModelMap)

		// Build tool registry for autonomous agents
		toolRegistry := tools.NewRegistry()
		tools.RegisterScheduleTools(toolRegistry, scheduleService)
		tools.RegisterProjectTools(toolRegistry, projectService, weatherService)
		tools.RegisterCommunicationTools(toolRegistry, directoryService, notificationService)
		tools.RegisterFeedTools(toolRegistry, feedService, agentActionService)

		agentRunner := agents.NewAgentRunner(claudeClient, toolRegistry)

		// Wire Claude reasoning to autonomous agents
		dailyFocusAgent.WithClaudeRunner(agentRunner)

		// Wire tool runner for executing approved agent actions
		actionRunner := tools.NewActionRunnerAdapter(toolRegistry)
		agentActionService.WithToolRunner(adapters.NewActionToolAdapter(actionRunner))

		log.Printf("INFO: Claude-powered agents ENABLED (%d tools registered)", len(toolRegistry.Definitions()))
	} else {
		log.Println("INFO: Claude-powered agents DISABLED (set ANTHROPIC_API_KEY to enable, falling back to Gemini)")
	}

	// Procurement Agent for long-lead item monitoring
	// See PRODUCTION_PLAN.md Step 46, 49
	// Config Decoupling: Load ProcurementConfig from environment (defaults if not set)
	procurementCfg := config.LoadProcurementConfigFromEnv()
	procurementAgent := agents.NewProcurementAgentWithDB(dbPool, weatherService, realClock, procurementCfg).
		WithFeedWriter(feedService) // V2 Feed: Write procurement cards

	// V2 Phase 7: Passive drift detection agent
	// See FRONTEND_V2_SPEC.md §11.2
	driftRepo := agents.NewPgDriftRepository(dbPool)
	driftAgent := agents.NewDriftDetectionAgent(driftRepo, realClock).
		WithFeedWriter(feedService)

	// 6.5 FutureShade Action Bridge: Initialize Skills Registry
	// See specs/FUTURESHADE_AGENTS_SPEC.md Section 4
	skillRegistry := skills.NewRegistry()
	skillRegistry.Register(skills.NewProcurementSyncSkill(procurementAgent))
	skillRegistry.Register(skills.NewDailyFocusSyncSkill(dailyFocusAgent))
	skillRegistry.Register(skills.NewScheduleRecalcSkill(scheduleService))

	// Initialize FutureShade config (defaults to disabled if not configured)
	futureShadeConfig := futureshade.Config{
		Enabled: cfg.FutureShadeEnabled,
	}
	if futureShadeConfig.Enabled {
		log.Println("INFO: FutureShade Action Bridge is ENABLED")
	} else {
		log.Println("INFO: FutureShade Action Bridge is DISABLED (set FUTURESHADE_ENABLED=true to enable)")
	}

	// Initialize Gateway Repository for execution logs
	executionRepo := gateway.NewRepository(dbPool)

	// Intelligence services for worker handlers (Features 1, 5, 6)
	delayCascadeService := service.NewDelayCascadeService(dbPool, feedService)
	calibrationService := service.NewCalibrationService(dbPool)
	resourceConflictService := service.NewResourceConflictService(dbPool, feedService)

	// 7. Initialize Worker Handlers
	// P1 Performance Fix: Pass db and clock for notification handler
	workerHandler := worker.NewWorkerHandler(dailyFocusAgent, procurementAgent, dbPool, realClock).
		WithSkillExecution(skillRegistry, executionRepo, futureShadeConfig).
		WithDriftDetection(driftAgent).
		WithAgentActionExpiry(agentActionService).
		WithBriefingNotification(notificationService, directoryService).
		WithDelayCascade(delayCascadeService).
		WithCalibration(calibrationService).
		WithResourceConflict(resourceConflictService)

	// 7.5 Automated PR Review: Initialize GitHub service and Tribunal integration
	// See docs/AUTOMATED_PR_REVIEW_PRD.md
	if cfg.GitHubPAT != "" && cfg.GitHubWebhookSecret != "" {
		if aiClient == nil {
			log.Println("WARN: Automated PR Review requires Vertex AI — DISABLED (aiClient is nil)")
		} else {
			log.Println("INFO: Automated PR Review is ENABLED")
			githubService := service.NewGitHubService(cfg.GitHubPAT)
			tribunalRepo := tribunal.NewRepository(dbPool)

			jury := tribunal.Jury{
				Coordinator: aiClient,
				Architect:   aiClient,
				Historian:   aiClient,
			}
			tribunalEngine := tribunal.NewConsensusEngine(jury, tribunalRepo)

			workerHandler = workerHandler.WithPRReview(githubService, tribunalEngine, tribunalRepo)
		}
	} else {
		log.Println("INFO: Automated PR Review is DISABLED (set GITHUB_PAT and GITHUB_WEBHOOK_SECRET to enable)")
	}

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

	// V2 Phase 7: Register Drift Detection at 07:00 AM UTC (after daily briefing)
	// See FRONTEND_V2_SPEC.md §11.2
	if _, err := scheduler.RegisterEntry("0 7 * * *", worker.NewDriftDetectionTask()); err != nil {
		log.Fatalf("could not register drift detection cron: %v", err)
	}

	// Human-in-the-loop: Expire stale pending actions at 23:00 UTC daily
	if _, err := scheduler.RegisterEntry("0 23 * * *", worker.NewExpireAgentActionsTask()); err != nil {
		log.Fatalf("could not register expire agent actions cron: %v", err)
	}

	// Feature 6: Weekly cross-project resource conflict scan (Monday 08:00 UTC)
	if _, err := scheduler.RegisterEntry("0 8 * * 1", worker.NewResourceConflictScanTask()); err != nil {
		log.Fatalf("could not register resource conflict scan cron: %v", err)
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
	// FutureShade Action Bridge: Register skill execution handler
	srv.RegisterHandlerFunc(worker.TypeSkillExecution, workerHandler.HandleSkillExecution)
	// Automated PR Review: Register PR review handler
	srv.RegisterHandlerFunc(worker.TypeReviewPR, workerHandler.HandleReviewPR)
	// V2 Phase 7: Passive drift detection
	srv.RegisterHandlerFunc(worker.TypeDriftDetection, workerHandler.HandleDriftDetection)
	// Human-in-the-loop: Expire stale pending agent actions
	srv.RegisterHandlerFunc(worker.TypeExpireAgentActions, workerHandler.HandleExpireAgentActions)
	// Feature 4: Daily briefing push notification
	srv.RegisterHandlerFunc(worker.TypeDailyBriefingNotification, workerHandler.HandleDailyBriefingNotification)
	// Feature 1: Predictive delay propagation
	srv.RegisterHandlerFunc(worker.TypeDelayCascade, workerHandler.HandleDelayCascade)
	// Feature 5: Calibrate org multipliers on project completion
	srv.RegisterHandlerFunc(worker.TypeCalibrateOnCompletion, workerHandler.HandleCalibrateOnCompletion)
	// Feature 6: Weekly resource conflict scan
	srv.RegisterHandlerFunc(worker.TypeResourceConflictScan, workerHandler.HandleResourceConflictScan)

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
		log.Printf("FATAL: %v", err)
	case <-quit:
		log.Println("Shutting down worker...")
	}
	scheduler.Shutdown()
	srv.Shutdown()
	log.Println("Worker stopped")
}
