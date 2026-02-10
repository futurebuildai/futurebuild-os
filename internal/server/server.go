package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/colton/futurebuild/internal/adapters"
	"github.com/colton/futurebuild/internal/agents"
	"github.com/colton/futurebuild/internal/api/handlers"
	"github.com/colton/futurebuild/internal/auth"
	"github.com/colton/futurebuild/internal/chat"
	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/futureshade"
	"github.com/colton/futurebuild/internal/futureshade/shadow"
	"github.com/colton/futurebuild/internal/futureshade/tribunal"
	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/readiness"
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
	PortalHandler          *handlers.PortalHandler          // See LAUNCH_PLAN.md P2: Field Portal
	PortalAuthHandler      *handlers.PortalAuthHandler      // Phase 12: Portal magic-link auth (separate from Clerk)
	PortalDashboardHandler *handlers.PortalDashboardHandler // Portal Dashboard API: authenticated contact endpoints
	GitHubWebhookHandler *handlers.GitHubWebhookHandler // See docs/AUTOMATED_PR_REVIEW_PRD.md
	ClerkWebhookHandler  *handlers.ClerkWebhookHandler  // See PHASE_12_PRD.md Step 80
	OnboardingHandler    *handlers.OnboardingHandler    // See PHASE_11_PRD.md Step 75: The Interrogator Agent
	InvoiceHandler       *handlers.InvoiceHandler       // See PHASE_13_PRD.md Step 82: Interactive Invoice
	AssetHandler         *handlers.AssetHandler         // See STEP_84_FIELD_FEEDBACK.md: Vision status
	ConfigHandler        *handlers.ConfigHandler        // See STEP_87_CONFIG_PERSISTENCE.md: Physics settings
	ScheduleHandler      *handlers.ScheduleHandler      // Phase 14: Gantt schedule data endpoint
	ThreadHandler        *handlers.ThreadHandler        // Thread support: conversation threads
	CompletionHandler    *handlers.CompletionHandler    // Project Completion: complete + report
	ReadinessHandler     *handlers.ReadinessHandler     // Integration readiness checks
	FeedHandler          *handlers.FeedHandler          // V2: Portfolio feed endpoint
	AuthMiddleware       *middleware.AuthMiddleware
	PortalRateLimiter    *middleware.IPRateLimiter       // Phase 12: Rate limiter for portal auth endpoints
	PublicRateLimiter    *middleware.IPRateLimiter       // L7: Rate limiter for public invite/portal action endpoints
}

func NewServer(db *pgxpool.Pool, cfg *config.Config, aiClient ai.Client) *Server {
	projectService := service.NewProjectService(db)
	threadService := service.NewThreadService(db)
	projectHandler := handlers.NewProjectHandler(projectService, threadService)

	// See PRODUCTION_PLAN.md Step 32
	scheduleService := service.NewScheduleService(db)
	taskHandler := handlers.NewTaskHandler(scheduleService)
	scheduleHandler := handlers.NewScheduleHandler(scheduleService) // Phase 14: Gantt endpoint
	threadHandler := handlers.NewThreadHandler(threadService)

	// See PRODUCTION_PLAN.md Step 37
	// See PRODUCTION_PLAN.md Step 37
	invoiceService := service.NewInvoiceService(db, aiClient, cfg)
	// See PRODUCTION_PLAN.md Step 41
	documentService := service.NewDocumentService(db, aiClient)
	documentHandler := handlers.NewDocumentHandler(invoiceService, documentService)

	// See PHASE_13_PRD.md Step 82: Interactive Invoice
	invoiceHandler := handlers.NewInvoiceHandler(invoiceService)

	// See STEP_84_FIELD_FEEDBACK.md: Vision status endpoint
	assetService := service.NewAssetService(db)
	assetHandler := handlers.NewAssetHandler(assetService)

	// See STEP_87_CONFIG_PERSISTENCE.md: Physics settings persistence
	configService := service.NewConfigService(db)
	configHandler := handlers.NewConfigHandler(configService)

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
	// Provider selection controlled by NOTIFICATION_PROVIDER env var:
	// - "bird": Bird (MessageBird) for both SMS and email
	// - "legacy": Resend for email, Twilio for SMS
	// - default: Console provider (development mode)
	// See LAUNCH_STRATEGY.md Task A3 and LAUNCH_PLAN.md P2.
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

	// V2: Portfolio feed service (created early for agent injection)
	feedService := service.NewFeedService(db)

	inboundProcessor := agents.NewInboundProcessor(
		db,
		directoryService, // Implements InboundContactLookup
		adapters.NewScheduleServiceAdapter(scheduleService, db), // Implements InboundProgressUpdater
		visionVerifier,
		clock.RealClock{},
	).WithFeedWriter(feedService) // V2 Feed: Write sub confirmation/delay cards
	webhookHandler := handlers.NewWebhookHandler(inboundProcessor, cfg.WebhookSecret)

	// Phase 12: Main app auth via Clerk; AuthHandler serves /auth/me only.
	authHandler := handlers.NewAuthHandler()
	authMiddleware := middleware.NewAuthMiddleware(cfg, db)

	// Phase 12: Portal contacts still use magic-link auth (separate from Clerk).
	// AuthService provides token generation, storage, and verification for portal.
	authService := service.NewAuthService(db, cfg)
	portalAuthHandler := handlers.NewPortalAuthHandler(authService, notificationService, cfg.BaseURL)
	portalRateLimiter := middleware.NewIPRateLimiter(rate.Every(12*time.Second), 2, cfg.TrustedProxies)
	publicRateLimiter := middleware.NewIPRateLimiter(rate.Every(5*time.Second), 5, cfg.TrustedProxies)

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
	// ClerkClient is nil when CLERK_SECRET_KEY is not set (Clerk user creation disabled).
	clerkClient := service.NewClerkClient(cfg.ClerkSecretKey)
	inviteService := service.NewInviteService(db, clerkClient)
	inviteHandler := handlers.NewInviteHandler(inviteService, notificationService, cfg.BaseURL)

	// See LAUNCH_PLAN.md User Profile Endpoint (P0)
	userService := service.NewUserService(db)
	userHandler := handlers.NewUserHandler(db, userService)

	// See LAUNCH_PLAN.md P2: Field Portal (Mobile)
	portalService := service.NewPortalService(db, notificationService, cfg.BaseURL)
	portalHandler := handlers.NewPortalHandler(portalService)
	portalDashboardHandler := handlers.NewPortalDashboardHandler(portalService)

	// See docs/AUTOMATED_PR_REVIEW_PRD.md: GitHub Webhook Handler
	// Only initialize if webhook secret is configured (fail-closed handled in handler)
	var githubWebhookHandler *handlers.GitHubWebhookHandler
	if cfg.GitHubWebhookSecret != "" {
		githubWebhookHandler = handlers.NewGitHubWebhookHandler(cfg.GitHubWebhookSecret, cfg.RedisURL)
	}

	// See PHASE_12_PRD.md Step 80: Clerk Webhook Handler for org/user sync
	// Only initialize if webhook secret is configured (fail-closed handled in handler)
	var clerkWebhookHandler *handlers.ClerkWebhookHandler
	if cfg.ClerkWebhookSecret != "" {
		clerkWebhookHandler = handlers.NewClerkWebhookHandler(db, cfg.ClerkWebhookSecret)
	}

	// V2: Portfolio feed handler (feedService created earlier for agent injection)
	feedHandler := handlers.NewFeedHandler(feedService)

	// Project Completion: service + handler
	completionService := service.NewCompletionService(db)
	completionHandler := handlers.NewCompletionHandler(completionService, notificationService, directoryService)

	// Integration Readiness Check System: per-provider probes for 3P service health.
	// Each probe creates a short-lived client from raw config — no interference with live services.
	// Notification probes are conditionally added based on NOTIFICATION_PROVIDER.
	readinessCheckers := []readiness.Checker{
		readiness.NewDatabaseProbe(db),
		readiness.NewClerkProbe(cfg.ClerkIssuerURL),
		readiness.NewRedisProbe(cfg.RedisURL),
		readiness.NewVertexAIProbe(cfg.VertexProjectID, cfg.VertexLocation),
		readiness.NewS3Probe(cfg.S3Endpoint, cfg.S3Bucket, cfg.S3AccessKey, cfg.S3SecretKey),
	}

	// Add notification probes based on provider selection
	switch cfg.NotificationProvider {
	case "bird":
		readinessCheckers = append(readinessCheckers, readiness.NewBirdProbe(cfg.BirdAccessKey))
	case "legacy":
		readinessCheckers = append(readinessCheckers,
			readiness.NewResendProbe(cfg.ResendAPIKey),
			readiness.NewTwilioProbe(cfg.TwilioAccountSID, cfg.TwilioAuthToken),
		)
	// default: no notification probes for console mode
	}

	readinessService := readiness.NewService(15*time.Second, readinessCheckers...)
	readinessHandler := handlers.NewReadinessHandler(readinessService, cfg.Environment)

	// See PHASE_11_PRD.md Step 75: The Interrogator Agent
	// C5 Fix: Guard against nil AI client (onboarding requires Gemini Vision API)
	var onboardingHandler *handlers.OnboardingHandler
	if aiClient != nil {
		interrogatorService := service.NewInterrogatorService(aiClient)
		onboardingHandler = handlers.NewOnboardingHandler(interrogatorService)
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
		PortalHandler:          portalHandler,          // See LAUNCH_PLAN.md P2: Field Portal
		PortalAuthHandler:      portalAuthHandler,      // Phase 12: Portal magic-link auth
		PortalDashboardHandler: portalDashboardHandler, // Portal Dashboard API
		GitHubWebhookHandler: githubWebhookHandler, // See docs/AUTOMATED_PR_REVIEW_PRD.md
		ClerkWebhookHandler:  clerkWebhookHandler,  // See PHASE_12_PRD.md Step 80
		OnboardingHandler:    onboardingHandler,    // See PHASE_11_PRD.md Step 75
		InvoiceHandler:       invoiceHandler,       // See PHASE_13_PRD.md Step 82
		AssetHandler:         assetHandler,         // See STEP_84_FIELD_FEEDBACK.md
		ConfigHandler:        configHandler,        // See STEP_87_CONFIG_PERSISTENCE.md
		ScheduleHandler:      scheduleHandler,      // Phase 14: Gantt schedule data
		ThreadHandler:        threadHandler,        // Thread support
		CompletionHandler:    completionHandler,    // Project Completion
		ReadinessHandler:     readinessHandler,     // Integration readiness checks
		FeedHandler:          feedHandler,          // V2: Portfolio feed
		AuthMiddleware:       authMiddleware,
		PortalRateLimiter:    portalRateLimiter,
		PublicRateLimiter:    publicRateLimiter,
	}

	s.routes()
	return s
}

func (s *Server) routes() {
	s.Router.Use(chiMiddleware.Logger)
	s.Router.Use(chiMiddleware.Recoverer)
	s.Router.Use(securityHeaders)

	s.Router.Get("/health", s.HandleHealth)

	s.Router.Route("/api/v1", func(r chi.Router) {
		// Phase 12: Legacy /auth/login and /auth/verify removed — Clerk handles sign-in.
		// See STEP_78_AUTH_PROVIDER.md Section 2.1
		r.Route("/auth", func(r chi.Router) {
			r.Use(s.AuthMiddleware.RequireAuth)
			r.Get("/me", s.AuthHandler.Me)
		})

		// See LAUNCH_STRATEGY.md Task B2: User Invite Flow
		// Public endpoints for accepting invitations (L7: rate-limited)
		r.Route("/invites", func(r chi.Router) {
			r.Use(middleware.RateLimit(s.PublicRateLimiter))
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
		// Public endpoints for one-time action links (L7: rate-limited)
		r.Route("/portal", func(r chi.Router) {
			r.Use(middleware.RateLimit(s.PublicRateLimiter))
			r.Get("/action/{token}", s.PortalHandler.HandleVerifyActionToken)
			r.Post("/action/{token}", s.PortalHandler.HandleSubmitAction)

			// Phase 12: Portal magic-link auth (separate from Clerk).
			// Contacts use email magic links, not Clerk SSO.
			r.Route("/auth", func(r chi.Router) {
				r.Use(middleware.RateLimit(s.PortalRateLimiter))
				r.Post("/login", s.PortalAuthHandler.Login)
				r.Get("/verify", s.PortalAuthHandler.Verify)
			})

			// Authenticated portal dashboard endpoints.
			// Protected by portal JWT middleware (HS256, subject_type=contact).
			r.Route("/me", func(r chi.Router) {
				r.Use(s.AuthMiddleware.RequirePortalAuth)

				r.Get("/projects", s.PortalDashboardHandler.ListProjects)

				r.Route("/projects/{id}", func(r chi.Router) {
					r.Get("/tasks", s.PortalDashboardHandler.ListProjectTasks)
					r.Get("/dependencies", s.PortalDashboardHandler.GetDependencies)

					r.Get("/messages", s.PortalDashboardHandler.ListMessages)
					r.Post("/messages", s.PortalDashboardHandler.SendMessage)

					r.Get("/documents", s.PortalDashboardHandler.ListDocuments)
					r.Post("/documents", s.PortalDashboardHandler.UploadDocument)

					r.Get("/invoices", s.PortalDashboardHandler.ListInvoices)
					r.Post("/invoices", s.PortalDashboardHandler.UploadInvoice)
				})
			})
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

		// Org member listing (Team page) — any authenticated user can view members
		r.Route("/org", func(r chi.Router) {
			r.Use(s.AuthMiddleware.RequireAuth)
			r.Get("/members", s.UserHandler.ListMembers)
		})

		// See STEP_87_CONFIG_PERSISTENCE.md: Org-level physics settings
		// C-1 Fix: GET requires ScopeProjectRead (all roles), PUT requires ScopeSettingsWrite (Admin/Builder)
		r.Route("/org/settings", func(r chi.Router) {
			r.Use(s.AuthMiddleware.RequireAuth)
			r.With(s.AuthMiddleware.RequirePermission(auth.ScopeProjectRead)).Get("/physics", s.ConfigHandler.GetPhysics)
			r.With(s.AuthMiddleware.RequirePermission(auth.ScopeSettingsWrite)).Put("/physics", s.ConfigHandler.UpdatePhysics)
		})

		// V2: Portfolio feed — aggregated feed cards across all projects
		// See FRONTEND_V2_SPEC.md §5.1
		r.Route("/portfolio", func(r chi.Router) {
			r.Use(s.AuthMiddleware.RequireAuth)
			r.With(s.AuthMiddleware.RequirePermission(auth.ScopeProjectRead)).Get("/feed", s.FeedHandler.GetFeed)
			r.With(s.AuthMiddleware.RequirePermission(auth.ScopeTaskWrite)).Post("/feed/action", s.FeedHandler.ExecuteAction)
			r.With(s.AuthMiddleware.RequirePermission(auth.ScopeTaskWrite)).Post("/feed/dismiss", s.FeedHandler.DismissCard)
			r.With(s.AuthMiddleware.RequirePermission(auth.ScopeTaskWrite)).Post("/feed/snooze", s.FeedHandler.SnoozeCard)
		})

		r.Route("/projects", func(r chi.Router) {
			r.Use(s.AuthMiddleware.RequireAuth) // L7 Security Fix: BOLA remediation
			// Step 81: Scope-based RBAC — write operations require specific permissions
			r.With(s.AuthMiddleware.RequirePermission(auth.ScopeProjectCreate)).Post("/", s.ProjectHandler.CreateProject)
			r.With(s.AuthMiddleware.RequirePermission(auth.ScopeProjectRead)).Get("/{id}", s.ProjectHandler.GetProject)
			r.With(s.AuthMiddleware.RequirePermission(auth.ScopeProjectRead)).Get("/{id}/procurement", s.ProjectHandler.GetProcurementItems)

			// Step 85: Project asset gallery with vision badges
			r.With(s.AuthMiddleware.RequirePermission(auth.ScopeBudgetRead)).Get("/{id}/assets", s.AssetHandler.ListProjectAssets)

			// Project Completion: complete project + get report
			r.With(s.AuthMiddleware.RequirePermission(auth.ScopeProjectComplete)).Post("/{id}/complete", s.CompletionHandler.CompleteProject)
			r.With(s.AuthMiddleware.RequirePermission(auth.ScopeProjectRead)).Get("/{id}/completion-report", s.CompletionHandler.GetCompletionReport)

			// Phase 14: Schedule/Gantt endpoint for frontend Gantt artifact
			r.Route("/{id}/schedule", func(r chi.Router) {
				r.With(s.AuthMiddleware.RequirePermission(auth.ScopeProjectRead)).Get("/", s.ScheduleHandler.GetSchedule)
				r.With(s.AuthMiddleware.RequirePermission(auth.ScopeTaskWrite)).Post("/recalculate", s.ScheduleHandler.RecalculateSchedule)
			})

			// Task endpoints - See PRODUCTION_PLAN.md Step 32
			r.Route("/{id}/tasks", func(r chi.Router) {
				r.With(s.AuthMiddleware.RequirePermission(auth.ScopeTaskWrite)).Put("/{task_id}", s.TaskHandler.UpdateTask)
				r.With(s.AuthMiddleware.RequirePermission(auth.ScopeTaskWrite)).Post("/{task_id}/progress", s.TaskHandler.RecordProgress)
				r.With(s.AuthMiddleware.RequirePermission(auth.ScopeTaskWrite)).Post("/{task_id}/inspection", s.TaskHandler.RecordInspection)
			})

			// Thread endpoints - conversation threads within projects
			r.Route("/{id}/threads", func(r chi.Router) {
				r.With(s.AuthMiddleware.RequirePermission(auth.ScopeChatRead)).Get("/", s.ThreadHandler.ListThreads)
				r.With(s.AuthMiddleware.RequirePermission(auth.ScopeChatWrite)).Post("/", s.ThreadHandler.CreateThread)
				r.With(s.AuthMiddleware.RequirePermission(auth.ScopeChatRead)).Get("/{threadId}", s.ThreadHandler.GetThread)
				r.With(s.AuthMiddleware.RequirePermission(auth.ScopeChatWrite)).Post("/{threadId}/archive", s.ThreadHandler.ArchiveThread)
				r.With(s.AuthMiddleware.RequirePermission(auth.ScopeChatWrite)).Post("/{threadId}/unarchive", s.ThreadHandler.UnarchiveThread)
				r.With(s.AuthMiddleware.RequirePermission(auth.ScopeChatRead)).Get("/{threadId}/messages", s.ThreadHandler.GetThreadMessages)
			})
		})
		// See STEP_84_FIELD_FEEDBACK.md: Vision analysis status polling
		// M1 Fix: Requires auth + budget:read scope (field users need to see analysis results)
		r.Route("/vision", func(r chi.Router) {
			r.Use(s.AuthMiddleware.RequireAuth)
			r.With(s.AuthMiddleware.RequirePermission(auth.ScopeBudgetRead)).Get("/status/{id}", s.AssetHandler.GetVisionStatus)
		})

		// See PHASE_13_PRD.md Steps 82-83: Invoice CRUD + Approval
		r.Route("/invoices", func(r chi.Router) {
			r.Use(s.AuthMiddleware.RequireAuth)
			r.With(s.AuthMiddleware.RequirePermission(auth.ScopeBudgetRead)).Get("/{id}", s.InvoiceHandler.GetInvoice)
			r.With(s.AuthMiddleware.RequirePermission(auth.ScopeFinanceEdit)).Put("/{id}", s.InvoiceHandler.UpdateInvoice)
			// Step 83: Approval requires budget:approve scope (Admin + explicit grant)
			r.With(s.AuthMiddleware.RequirePermission(auth.ScopeBudgetApprove)).Post("/{id}/approve", s.InvoiceHandler.ApproveInvoice)
			r.With(s.AuthMiddleware.RequirePermission(auth.ScopeBudgetApprove)).Post("/{id}/reject", s.InvoiceHandler.RejectInvoice)
		})

		r.Route("/documents", func(r chi.Router) {
			r.Use(s.AuthMiddleware.RequireAuth) // L7 Security Fix: BOLA remediation
			// Step 81: Document write operations require document:write scope
			r.With(s.AuthMiddleware.RequirePermission(auth.ScopeDocumentWrite)).Post("/analyze", s.DocumentHandler.AnalyzeDocument)
			// See PRODUCTION_PLAN.md Step 41
			r.With(s.AuthMiddleware.RequirePermission(auth.ScopeDocumentWrite)).Post("/{id}/reprocess", s.DocumentHandler.ReprocessDocument)
		})

		// See PRODUCTION_PLAN.md Step 43.5: Chat endpoint with Auth
		r.Route("/chat", func(r chi.Router) {
			r.Use(s.AuthMiddleware.RequireAuth)
			// Step 81: Chat write requires chat:write scope (Viewers can read but not send)
			r.With(s.AuthMiddleware.RequirePermission(auth.ScopeChatWrite)).Post("/", s.ChatHandler.HandleChat)
		})

		// See PHASE_11_PRD.md Step 75: Interrogator Agent
		// L7: Auth required, rate-limited to prevent AI API abuse
		// C5 Fix: Only register route if AI client is configured
		r.Route("/agent", func(r chi.Router) {
			r.Use(s.AuthMiddleware.RequireAuth)
			// FUTURE: Add rate limiter when agent abuse detected
			// r.Use(middleware.RateLimit(s.AgentRateLimiter))
			if s.OnboardingHandler != nil {
				r.Post("/onboard", s.OnboardingHandler.HandleOnboard)
			}
		})

		// See PRODUCTION_PLAN.md Step 48: Inbound Webhook Endpoints
		r.Route("/webhooks", func(r chi.Router) {
			r.Post("/sms", s.WebhookHandler.HandleSMS)
			r.Post("/email", s.WebhookHandler.HandleEmail)
			// See docs/AUTOMATED_PR_REVIEW_PRD.md: GitHub PR Review Webhook
			if s.GitHubWebhookHandler != nil {
				r.Post("/github", s.GitHubWebhookHandler.HandleGitHubWebhook)
			}
			// See PHASE_12_PRD.md Step 80: Clerk org/user sync webhook
			if s.ClerkWebhookHandler != nil {
				r.Post("/clerk", s.ClerkWebhookHandler.HandleClerkWebhook)
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

		// Integration Readiness: deep health checks for all 3P services (Admin only).
		// Unlike /health (DB-only, polled by DO every 10s), this runs probes against
		// Clerk, Redis, Resend, Twilio, Vertex AI, and S3.
		r.Route("/readiness", func(r chi.Router) {
			r.Use(s.AuthMiddleware.RequireAuth)
			r.Use(s.AuthMiddleware.RequireRole(types.UserRoleAdmin))
			r.Get("/", s.ReadinessHandler.HandleReadiness)
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
	// Vite outputs to dist/assets/, so we serve the entire dist dir and let
	// the /assets/* route map directly to dist/assets/*
	fileServer := http.FileServer(http.Dir(staticDir))
	s.Router.Handle("/assets/*", fileServer)

	// SPA catch-all: serve index.html for all non-API routes
	s.Router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, staticDir+"/index.html")
	})
}

// NewHTTPServer creates an http.Server with production-safe timeouts.
// The caller is responsible for calling ListenAndServe and Shutdown.
// See Staging Readiness Audit: Findings #1 (Graceful Shutdown) and #2 (Request Timeouts).
func (s *Server) NewHTTPServer() *http.Server {
	addr := fmt.Sprintf(":%d", s.Cfg.AppPort)
	return &http.Server{
		Addr:              addr,
		Handler:           s.Router,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

// Start is a convenience method that blocks until the server exits.
// For production use, prefer NewHTTPServer() + signal handling in main().
func (s *Server) Start() error {
	httpServer := s.NewHTTPServer()
	fmt.Printf("Server starting on %s\n", httpServer.Addr)
	return httpServer.ListenAndServe()
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

// securityHeaders adds standard HTTP security headers to every response.
// See OWASP Secure Headers Project.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}
