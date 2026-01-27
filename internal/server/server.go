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
	"github.com/colton/futurebuild/internal/futureshade"
	"github.com/colton/futurebuild/internal/futureshade/shadow"
	"github.com/colton/futurebuild/internal/futureshade/tribunal"
	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/colton/futurebuild/pkg/clock"
	"github.com/colton/futurebuild/pkg/types"
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
	ChatHandler       *handlers.ChatHandler       // See PRODUCTION_PLAN.md Step 43.5
	WebhookHandler    *handlers.WebhookHandler    // See PRODUCTION_PLAN.md Step 48
	FutureShadeHandler *handlers.FutureShadeHandler // See FUTURESHADE_INIT_specs.md
	TribunalHandler    *handlers.TribunalHandler    // See SHADOW_VIEWER_specs.md
	ShadowHandler      *handlers.ShadowHandler      // See SHADOW_VIEWER_specs.md
	InviteHandler        *handlers.InviteHandler        // See LAUNCH_STRATEGY.md Task B2
	UserHandler          *handlers.UserHandler          // See LAUNCH_PLAN.md User Profile Endpoint
	PortalHandler        *handlers.PortalHandler        // See LAUNCH_PLAN.md P2: Field Portal
	GitHubWebhookHandler *handlers.GitHubWebhookHandler // See docs/AUTOMATED_PR_REVIEW_PRD.md
	AuthMiddleware       *middleware.AuthMiddleware
	AuthRateLimiter      *middleware.IPRateLimiter
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

	// Initialize Audit WAL for durability fallback.
	// In production/staging, use FileAuditWAL. In development, use NoOp for simplicity.
	// See LAUNCH_PLAN.md Production Safety section.
	var wal chat.AuditWAL
	if cfg.Environment == "production" || cfg.Environment == "staging" {
		fileWAL, err := chat.NewFileAuditWAL(cfg.AuditWALPath)
		if err != nil {
			panic(fmt.Sprintf("CRITICAL: Failed to initialize AuditWAL at %s: %v", cfg.AuditWALPath, err))
		}
		wal = fileWAL
		fmt.Printf("Audit WAL: FileAuditWAL at %s\n", cfg.AuditWALPath)
	} else {
		wal = &chat.NoOpAuditWAL{}
		fmt.Println("Audit WAL: NoOp (development mode)")
	}

	// Initialize Circuit Breaker for audit system availability.
	// Uses in-memory SimpleCircuitBreaker with sensible defaults.
	// See LAUNCH_PLAN.md Production Safety section.
	circuitBreaker := chat.NewSimpleCircuitBreaker(chat.DefaultCircuitBreakerConfig())
	fmt.Printf("Circuit Breaker: SimpleCircuitBreaker (threshold=%d, timeout=%s)\n",
		chat.DefaultCircuitBreakerConfig().FailureThreshold,
		chat.DefaultCircuitBreakerConfig().OpenTimeout)

	chatOrchestrator, err := chat.NewOrchestrator(
		messageStore, scheduleService, scheduleService, invoiceService, dlq,
		wal,
		circuitBreaker,
	)
	if err != nil {
		panic(fmt.Sprintf("CRITICAL: Failed to initialize Chat Orchestrator: %v", err))
	}
	chatHandler := handlers.NewChatHandler(chatOrchestrator)

	// Initialize notification service based on environment.
	// Uses composite provider: SendGrid for email, Twilio for SMS.
	// Falls back to Console provider in development mode.
	// See LAUNCH_STRATEGY.md Task A3 and LAUNCH_PLAN.md P2.
	notificationService := service.NewNotificationService(
		cfg.SendGridAPIKey,
		cfg.EmailFromAddress,
		cfg.EmailFromName,
		cfg.TwilioAccountSID,
		cfg.TwilioAuthToken,
		cfg.TwilioFromNumber,
	)
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
	authHandler := handlers.NewAuthHandler(authService, notificationService, cfg.BaseURL)
	authMiddleware := middleware.NewAuthMiddleware(cfg)

	authRateLimiter := middleware.NewIPRateLimiter(rate.Every(12*time.Second), 2, cfg.TrustedProxies)

	// See FUTURESHADE_INIT_specs.md: Initialize FutureShade with Fail Open strategy.
	// If configuration is missing, the service returns a disabled NoOp instance.
	futureShadeConfig := &futureshade.Config{
		Enabled: cfg.FutureShadeEnabled,
		APIKey:  cfg.FutureShadeAPIKey,
		ModelID: cfg.FutureShadeModelID,
	}
	futureShadeService := futureshade.NewService(futureShadeConfig)
	futureShadeHandler := handlers.NewFutureShadeHandler(futureShadeService)

	// See SHADOW_VIEWER_specs.md: Tribunal and ShadowDocs handlers
	tribunalRepo := tribunal.NewRepository(db)
	tribunalHandler := handlers.NewTribunalHandler(tribunalRepo)
	shadowService := shadow.NewDocsService(cfg.ProjectRoot)
	shadowHandler := handlers.NewShadowHandler(shadowService)

	// See LAUNCH_STRATEGY.md Task B2: User Invite Flow
	inviteService := service.NewInviteService(db)
	inviteHandler := handlers.NewInviteHandler(inviteService, notificationService, cfg.BaseURL)

	// See LAUNCH_PLAN.md User Profile Endpoint (P0)
	userHandler := handlers.NewUserHandler(db)

	// See LAUNCH_PLAN.md P2: Field Portal (Mobile)
	portalService := service.NewPortalService(db, notificationService, cfg.BaseURL)
	portalHandler := handlers.NewPortalHandler(portalService)

	// See docs/AUTOMATED_PR_REVIEW_PRD.md: GitHub Webhook Handler
	// Only initialize if webhook secret is configured (fail-closed handled in handler)
	var githubWebhookHandler *handlers.GitHubWebhookHandler
	if cfg.GitHubWebhookSecret != "" {
		githubWebhookHandler = handlers.NewGitHubWebhookHandler(cfg.GitHubWebhookSecret, cfg.RedisURL)
	}

	s := &Server{
		Router:          chi.NewRouter(),
		DB:              db,
		Cfg:             cfg,
		ProjectHandler:  projectHandler,
		TaskHandler:     taskHandler,
		AuthHandler:     authHandler,
		DocumentHandler: documentHandler,
		ChatHandler:        chatHandler,        // See PRODUCTION_PLAN.md Step 43.5
		WebhookHandler:     webhookHandler,     // See PRODUCTION_PLAN.md Step 48
		FutureShadeHandler: futureShadeHandler, // See FUTURESHADE_INIT_specs.md
		TribunalHandler:    tribunalHandler,    // See SHADOW_VIEWER_specs.md
		ShadowHandler:      shadowHandler,      // See SHADOW_VIEWER_specs.md
		InviteHandler:        inviteHandler,        // See LAUNCH_STRATEGY.md Task B2
		UserHandler:          userHandler,          // See LAUNCH_PLAN.md User Profile Endpoint
		PortalHandler:        portalHandler,        // See LAUNCH_PLAN.md P2: Field Portal
		GitHubWebhookHandler: githubWebhookHandler, // See docs/AUTOMATED_PR_REVIEW_PRD.md
		AuthMiddleware:       authMiddleware,
		AuthRateLimiter:      authRateLimiter,
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

		// See LAUNCH_STRATEGY.md Task B2: User Invite Flow
		// Public endpoints for accepting invitations
		r.Route("/invites", func(r chi.Router) {
			r.Get("/info", s.InviteHandler.GetInviteInfo) // Public: get invite info by token
			r.Post("/accept", s.InviteHandler.AcceptInvite) // Public: accept invite and create account
		})

		// Admin endpoints for managing invitations
		r.Route("/admin/invites", func(r chi.Router) {
			r.Use(s.AuthMiddleware.RequireAuth)
			r.Use(s.AuthMiddleware.RequireRole(types.UserRoleAdmin))
			r.Post("/", s.InviteHandler.CreateInvite)
			r.Get("/", s.InviteHandler.ListInvites)
			r.Delete("/{id}", s.InviteHandler.RevokeInvite)
		})

		// See LAUNCH_PLAN.md P2: Field Portal (Mobile)
		// Public endpoints for one-time action links
		r.Route("/portal", func(r chi.Router) {
			r.Get("/action/{token}", s.PortalHandler.HandleVerifyActionToken)
			r.Post("/action/{token}", s.PortalHandler.HandleSubmitAction)
		})

		// Admin endpoint for creating action links
		r.Route("/admin/portal", func(r chi.Router) {
			r.Use(s.AuthMiddleware.RequireAuth)
			r.Use(s.AuthMiddleware.RequireRole(types.UserRoleAdmin))
			r.Post("/link", s.PortalHandler.HandleCreateActionLink)
		})

		// See LAUNCH_PLAN.md User Profile Endpoint (P0)
		r.Route("/users", func(r chi.Router) {
			r.Use(s.AuthMiddleware.RequireAuth)
			r.Get("/me", s.UserHandler.GetProfile)
			r.Put("/me", s.UserHandler.UpdateProfile)
		})

		r.Route("/projects", func(r chi.Router) {
			r.Use(s.AuthMiddleware.RequireAuth) // L7 Security Fix: BOLA remediation
			r.Post("/", s.ProjectHandler.CreateProject)
			r.Get("/{id}", s.ProjectHandler.GetProject)
			r.Get("/{id}/procurement", s.ProjectHandler.GetProcurementItems)

			// Task endpoints - See PRODUCTION_PLAN.md Step 32
			r.Route("/{id}/tasks", func(r chi.Router) {
				r.Put("/{task_id}", s.TaskHandler.UpdateTask)
				r.Post("/{task_id}/progress", s.TaskHandler.RecordProgress)
				r.Post("/{task_id}/inspection", s.TaskHandler.RecordInspection)
			})
		})
		r.Route("/documents", func(r chi.Router) {
			r.Use(s.AuthMiddleware.RequireAuth) // L7 Security Fix: BOLA remediation
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
			// See docs/AUTOMATED_PR_REVIEW_PRD.md: GitHub PR Review Webhook
			if s.GitHubWebhookHandler != nil {
				r.Post("/github", s.GitHubWebhookHandler.HandleGitHubWebhook)
			}
		})

		// See FUTURESHADE_INIT_specs.md Section 3.2: FutureShade endpoints (Admin only)
		r.Route("/futureshade", func(r chi.Router) {
			r.Use(s.AuthMiddleware.RequireAuth)
			r.Use(s.AuthMiddleware.RequireRole(types.UserRoleAdmin))
			r.Get("/health", s.FutureShadeHandler.HandleHealth)
		})

		// See SHADOW_VIEWER_specs.md Section 3.1: Tribunal endpoints (Admin only)
		r.Route("/tribunal", func(r chi.Router) {
			r.Use(s.AuthMiddleware.RequireAuth)
			r.Use(s.AuthMiddleware.RequireRole(types.UserRoleAdmin))
			r.Get("/decisions", s.TribunalHandler.ListDecisions)
			r.Get("/decisions/{id}", s.TribunalHandler.GetDecision)
		})

		// See SHADOW_VIEWER_specs.md Section 3.2: ShadowDocs endpoints (Admin only)
		r.Route("/shadow", func(r chi.Router) {
			r.Use(s.AuthMiddleware.RequireAuth)
			r.Use(s.AuthMiddleware.RequireRole(types.UserRoleAdmin))
			r.Get("/docs/tree", s.ShadowHandler.GetTree)
			r.Get("/docs/content", s.ShadowHandler.GetContent)
		})
	})

	// Legacy webhook endpoint for backwards compatibility
	// Deprecated: Use /api/v1/webhooks/sms or /api/v1/webhooks/email
	s.Router.Route("/webhooks", func(r chi.Router) {
		r.Post("/messages", s.WebhookHandler.HandleInboundMessage)
	})

	// Serve frontend static files
	// Files are served from /app/frontend/dist in the container
	staticDir := "/app/frontend/dist"
	if s.Cfg.Environment == "development" || s.Cfg.Environment == "dev" {
		staticDir = "frontend/dist"
	}

	// Serve static assets (JS, CSS, images)
	fileServer := http.FileServer(http.Dir(staticDir))
	s.Router.Handle("/assets/*", http.StripPrefix("/assets/", fileServer))

	// SPA catch-all: serve index.html for all non-API routes
	s.Router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, staticDir+"/index.html")
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
