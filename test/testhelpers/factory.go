package testhelpers

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/colton/futurebuild/internal/agents"
	"github.com/colton/futurebuild/internal/api/handlers"
	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/internal/service/mocks"
	"github.com/colton/futurebuild/internal/worker"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/colton/futurebuild/pkg/clock"
	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// IntegrationStack holds the wired application components.
// L7 Quality: Thread-safe, properly cleaned up via t.Cleanup().
type IntegrationStack struct {
	DB               *pgxpool.Pool
	Router           *chi.Mux
	ProjectService   *service.ProjectService
	TaskService      *service.ScheduleService // Using ScheduleService as TaskService
	TaskHandler      *handlers.TaskHandler
	AsynqClient      *asynq.Client
	WorkerServer     *worker.Server
	ProcurementAgent *agents.ProcurementAgent
	RedisClient      *redis.Client
}

// NewIntegrationStack spins up containers, migrates DB, and wires the app.
// It registers cleanup on the testing.T.
// L7 Quality: All resources are properly cleaned up; goroutines are safe.
func NewIntegrationStack(t *testing.T) *IntegrationStack {
	// 1. Start Postgres Container
	dbPool, pgCleanup := StartPostgresContainer(t)
	t.Cleanup(pgCleanup)

	// 2. Start Redis Container
	redisAddr, redisCleanup := StartRedisContainer(t)
	t.Cleanup(redisCleanup)

	// 3. Setup Async Infrastructure
	// We use the real enqueuer which connects to the Redis container
	hydrationEnqueuer := worker.NewAsynqHydrationEnqueuer(redisAddr)
	t.Cleanup(func() { hydrationEnqueuer.Close() })

	// 4. Wire Services
	// ProjectService gets the hydration enqueuer (Event-Driven Trigger)
	projectService := service.NewProjectServiceWithHydration(dbPool, hydrationEnqueuer)
	projectHandler := handlers.NewProjectHandler(projectService)

	// ScheduleService (handles Task logic)
	scheduleService := service.NewScheduleService(dbPool)
	taskHandler := handlers.NewTaskHandler(scheduleService)

	// 5. Setup Worker Agents (Real Logic with Mocks for external IO)
	// NOTE: PgProjectRepository wraps ProjectService, not DB
	projRepo := agents.NewPgProjectRepository(projectService)
	procRepo := agents.NewPgProcurementRepository(dbPool)

	// Mocks for external services
	mockWeather := &mocks.MockWeatherService{}
	mockNotifier := &mocks.MockNotificationService{}
	mockAI := &ai.MockClient{}
	geocoder := service.NewGeocodingService() // Use real stub
	mockDir := &mocks.MockDirectoryService{}

	realClock := clock.RealClock{}

	// Agents
	procAgent := agents.NewProcurementAgent(procRepo, mockWeather, realClock, config.ProcurementConfig{})
	procAgent = procAgent.WithNotificationEnqueuer(mockNotifier)

	focusAgent := agents.NewDailyFocusAgent(
		projRepo,
		scheduleService,
		mockWeather,
		mockNotifier,
		mockAI,
		realClock,
		geocoder,
		mockDir,
	)

	// 6. Setup Worker Server
	workerHandler := worker.NewWorkerHandler(focusAgent, procAgent, dbPool, realClock)
	// Default testing queues
	queues := map[string]int{"critical": 6, "default": 3, "low": 1}
	workerServer := worker.NewServer(redisAddr, 10, queues)

	// Register Handlers
	workerServer.RegisterHandlerFunc(worker.TypeHydrateProject, workerHandler.HandleHydrateProject)
	workerServer.RegisterHandlerFunc(worker.TypeProcurementCheck, workerHandler.HandleProcurementCheck)
	workerServer.RegisterHandlerFunc(worker.TypeDailyBriefing, workerHandler.HandleDailyBriefing)
	workerServer.RegisterHandlerFunc(worker.TypeProcurementNotification, workerHandler.HandleProcurementNotification)

	// L7 Quality: Use log.Printf instead of t.Logf in goroutine to avoid data race
	// after test completion. The goroutine may outlive the test briefly during shutdown.
	go func() {
		if err := workerServer.Run(); err != nil {
			log.Printf("Worker server stopped: %v", err)
		}
	}()
	t.Cleanup(func() {
		workerServer.Shutdown()
	})

	// 7. Wire Router
	r := chi.NewRouter()
	r.Post("/projects", projectHandler.CreateProject)
	r.Get("/projects/{id}", projectHandler.GetProject)
	r.Get("/projects/{id}/procurement", projectHandler.GetProcurementItems)

	// Expose client for direct inspection if needed
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	t.Cleanup(func() { client.Close() })

	// Expose raw Redis client for flushing
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	t.Cleanup(func() { rdb.Close() })

	return &IntegrationStack{
		DB:               dbPool,
		Router:           r,
		ProjectService:   projectService,
		TaskService:      scheduleService,
		TaskHandler:      taskHandler,
		AsynqClient:      client,
		WorkerServer:     workerServer,
		ProcurementAgent: procAgent,
		RedisClient:      rdb,
	}
}

// TruncateAll cleans the database and redis between tests.
// L7 Quality: Uses quoted identifiers to prevent SQL injection.
func (s *IntegrationStack) TruncateAll(ctx context.Context) error {
	// 1. Truncate Postgres
	query := `
		SELECT tablename FROM pg_tables 
		WHERE schemaname = 'public' 
		  AND tablename NOT LIKE 'pg_%'
		  AND tablename != 'schema_migrations'
		  AND tablename NOT LIKE 'wbs_%'
	`
	rows, err := s.DB.Query(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		tables = append(tables, name)
	}

	if len(tables) > 0 {
		// L7 Quality: Use quoted identifiers to prevent SQL injection
		// even though table names come from pg_tables (trusted source).
		truncateSQL := fmt.Sprintf("TRUNCATE TABLE %q", tables[0])
		for _, t := range tables[1:] {
			truncateSQL += fmt.Sprintf(", %q", t)
		}
		truncateSQL += " CASCADE"

		_, err = s.DB.Exec(ctx, truncateSQL)
		if err != nil {
			return err
		}
	}

	// 2. Flush Redis (Isolation)
	if err := s.RedisClient.FlushAll(ctx).Err(); err != nil {
		return err
	}

	return nil
}
