package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/colton/futurebuild/internal/adapters"
	"github.com/colton/futurebuild/internal/agents"
	"github.com/colton/futurebuild/internal/api/handlers"
	"github.com/colton/futurebuild/internal/chat"
	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/colton/futurebuild/pkg/clock"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/time/rate"
)

type Server struct {
	Router *chi.Mux
	DB     *pgxpool.Pool
	Cfg    *config.Config

	ProjectHandler  *handlers.ProjectHandler
	TaskHandler     *handlers.TaskHandler
	AuthHandler     *handlers.AuthHandler
	DocumentHandler *handlers.DocumentHandler
	ChatHandler     *handlers.ChatHandler    // See PRODUCTION_PLAN.md Step 43.5
	WebhookHandler  *handlers.WebhookHandler // See PRODUCTION_PLAN.md Step 48
	AuthMiddleware  *middleware.AuthMiddleware
	AuthRateLimiter *middleware.IPRateLimiter
}

func NewServer(db *pgxpool.Pool, cfg *config.Config, aiClient ai.Client) *Server {
	projectService := service.NewProjectService(db)
	projectHandler := handlers.NewProjectHandler(projectService)

	// See PRODUCTION_PLAN.md Step 32
	scheduleService := service.NewScheduleService(db)
	taskHandler := handlers.NewTaskHandler(scheduleService)

	// See PRODUCTION_PLAN.md Step 37
	// See PRODUCTION_PLAN.md Step 37
	invoiceService := service.NewInvoiceService(db, aiClient, cfg)
	// See PRODUCTION_PLAN.md Step 41
	documentService := service.NewDocumentService(db, aiClient)
	documentHandler := handlers.NewDocumentHandler(invoiceService, documentService)

	// See PRODUCTION_PLAN.md Step 43.5: Chat Orchestrator wiring
	// ENGINEERING STANDARD: Instantiate MessagePersister explicitly, then inject.
	// ScheduleService satisfies both TaskService and ScheduleService interfaces.
	// P0 FIX: DLQ is now MANDATORY for compliance audit trails.
	messageStore := chat.NewPgxMessageStore(db)
	dlq := chat.NewAsynqDLQ(cfg.RedisURL)

	wal := &chat.NoOpAuditWAL{}                  // TODO: Replace with production WAL
	circuitBreaker := &chat.NoOpCircuitBreaker{} // TODO: Replace with production circuit breaker

	// PRODUCTION SAFETY CHECK (Code Review Issue 3B)
	// NoOp implementations are placeholders that skip critical functionality.
	// They MUST be replaced before production deployment.
	if cfg.Environment == "production" {
		// P0 FIX: Fail Fast in Production
		if _, ok := interface{}(wal).(*chat.NoOpAuditWAL); ok {
			panic("CRITICAL SECURITY ERROR: NoOpAuditWAL is not allowed in production! Configure a real WAL.")
		}
		if _, ok := interface{}(circuitBreaker).(*chat.NoOpCircuitBreaker); ok {
			panic("CRITICAL SECURITY ERROR: NoOpCircuitBreaker is not allowed in production! Configure a real CircuitBreaker.")
		}
	}

	chatOrchestrator, err := chat.NewOrchestrator(
		messageStore, scheduleService, scheduleService, invoiceService, dlq,
		wal,
		circuitBreaker,
	)
	if err != nil {
		panic(fmt.Sprintf("CRITICAL: Failed to initialize Chat Orchestrator: %v", err))
	}
	chatHandler := handlers.NewChatHandler(chatOrchestrator)

	notificationService := service.NewConsoleEmailProvider()
	directoryService := service.NewDirectoryService(db)

	// NOTE: Background agents (SubLiaisonAgent, DailyFocusAgent, ProcurementAgent) run in
	// worker process (cmd/worker/main.go), NOT the API server. Do not instantiate here.
	// See PRODUCTION_PLAN.md Steps 45-49

	// See PRODUCTION_PLAN.md Step 48: Inbound Processor (inbound message handling)
	// VisionService is optional - pass nil if AI client not configured
	// Technical Debt Remediation (P2): Uses adapters package for cleaner separation
	var visionVerifier agents.InboundVisionVerifier
	if aiClient != nil {
		visionService := service.NewVisionService(aiClient)
		visionVerifier = adapters.NewVisionServiceAdapter(visionService)
	}

	inboundProcessor := agents.NewInboundProcessor(
		db,
		directoryService, // Implements InboundContactLookup
		adapters.NewScheduleServiceAdapter(scheduleService, db), // Implements InboundProgressUpdater
		visionVerifier,
		clock.RealClock{},
	)
	webhookHandler := handlers.NewWebhookHandler(inboundProcessor, cfg.WebhookSecret)

	authService := service.NewAuthService(db, cfg)
	authHandler := handlers.NewAuthHandler(authService, notificationService, "http://localhost:8080")
	authMiddleware := middleware.NewAuthMiddleware(cfg)

	authRateLimiter := middleware.NewIPRateLimiter(rate.Every(12*time.Second), 2, cfg.TrustedProxies)

	s := &Server{
		Router:          chi.NewRouter(),
		DB:              db,
		Cfg:             cfg,
		ProjectHandler:  projectHandler,
		TaskHandler:     taskHandler,
		AuthHandler:     authHandler,
		DocumentHandler: documentHandler,
		ChatHandler:     chatHandler,    // See PRODUCTION_PLAN.md Step 43.5
		WebhookHandler:  webhookHandler, // See PRODUCTION_PLAN.md Step 48
		AuthMiddleware:  authMiddleware,
		AuthRateLimiter: authRateLimiter,
	}

	s.routes()
	return s
}

func (s *Server) routes() {
	s.Router.Use(chiMiddleware.Logger)
	s.Router.Use(chiMiddleware.Recoverer)

	s.Router.Get("/health", s.HandleHealth)

	s.Router.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Use(middleware.RateLimit(s.AuthRateLimiter))
			r.Post("/login", s.AuthHandler.Login)
			r.Get("/verify", s.AuthHandler.Verify)
		})
		r.Route("/projects", func(r chi.Router) {
			r.Post("/", s.ProjectHandler.CreateProject)
			r.Get("/{id}", s.ProjectHandler.GetProject)

			// Task endpoints - See PRODUCTION_PLAN.md Step 32
			r.Route("/{id}/tasks", func(r chi.Router) {
				r.Put("/{task_id}", s.TaskHandler.UpdateTask)
				r.Post("/{task_id}/progress", s.TaskHandler.RecordProgress)
				r.Post("/{task_id}/inspection", s.TaskHandler.RecordInspection)
			})
		})
		r.Route("/documents", func(r chi.Router) {
			r.Post("/analyze", s.DocumentHandler.AnalyzeDocument)
			// See PRODUCTION_PLAN.md Step 41
			r.Post("/{id}/reprocess", s.DocumentHandler.ReprocessDocument)
		})

		// See PRODUCTION_PLAN.md Step 43.5: Chat endpoint with Auth
		r.Route("/chat", func(r chi.Router) {
			r.Use(s.AuthMiddleware.RequireAuth)
			r.Post("/", s.ChatHandler.HandleChat)
		})

		// See PRODUCTION_PLAN.md Step 48: Inbound Webhook Endpoints
		r.Route("/webhooks", func(r chi.Router) {
			r.Post("/sms", s.WebhookHandler.HandleSMS)
			r.Post("/email", s.WebhookHandler.HandleEmail)
		})
	})

	// Legacy webhook endpoint for backwards compatibility
	// Deprecated: Use /api/v1/webhooks/sms or /api/v1/webhooks/email
	s.Router.Route("/webhooks", func(r chi.Router) {
		r.Post("/messages", s.WebhookHandler.HandleInboundMessage)
	})
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.Cfg.AppPort)
	fmt.Printf("Server starting on %s\n", addr)
	return http.ListenAndServe(addr, s.Router)
}

func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	err := s.DB.Ping(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("Database unavailable"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
