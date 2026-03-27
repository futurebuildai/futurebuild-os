# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

FutureBuild is an AI-powered construction project management platform. It uses the Residential Construction Path Model (CPM-res1.0) to automate scheduling for residential construction projects, starting from Permit Issued (WBS 5.2).

## Technology Stack

| Layer | Technology |
|-------|------------|
| Backend | Go 1.24+, Chi Router, PostgreSQL 15+ (pgvector), Redis (Asynq) |
| Frontend | Lit 3.0, TypeScript 5.0+ (Strict Mode), Vite, Signals (@lit-labs/preact-signals) |
| AI | Google Vertex AI (Gemini 2.5 Flash/Pro) + Anthropic Claude Opus (chat orchestration) |
| Auth | Clerk (main app), Magic link email (portal contacts), JWT tokens |
| Observability | OpenTelemetry (OTLP/HTTP) |
| Testing | Testcontainers (PostgreSQL, Redis) for integration tests |

**Hard Constraints:**
- NO React, NO ORMs (use raw SQL/pgx), NO Python logic (Go only)
- Database is the source of truth; agents are stateless calculators
- All TypeScript must compile with `noImplicitAny` enabled
- Frontend uses `exactOptionalPropertyTypes`, `noUncheckedIndexedAccess` — strictest TS settings

## Build & Development Commands

```bash
# Start full dev environment (Go API + Vite frontend concurrently)
npm run dev

# Backend only
go run ./cmd/api

# Frontend only
npm --prefix frontend run dev

# Audit (lint + type check — runs go vet + frontend build)
make audit

# Unit tests (excludes integration)
make test

# Run a single Go test
go test -v ./internal/chat/... -run TestOrchestrator

# Integration tests (requires running DB, uses testcontainers-go)
make test-integration

# Contract validation (Go/TS type parity)
make contract-test

# Frontend lint/format
npm --prefix frontend run lint
npm --prefix frontend run lint:fix
npm --prefix frontend run format
npm --prefix frontend run format:check

# Database migrations (golang-migrate, 6-digit zero-padded sequential)
make migrate-up
make migrate-down
make migrate-create name=<migration_name>

# Background worker (Asynq cron jobs)
make run-worker

# Seed demo data
make seed-demo
```

## Entry Points

| Binary | Path | Purpose |
|--------|------|---------|
| API Server | `cmd/api/main.go` | HTTP server, Chi router, all API endpoints |
| Worker | `cmd/worker/main.go` | Asynq cron jobs (daily briefings, procurement, drift detection) |
| Demo Seed | `cmd/seed-demo/main.go` | Idempotent demo data seeder |

The API server supports `--readiness-check` flag for CI/CD health probes (runs probes and exits).

## Architecture

### Backend (`internal/`)

- **`server/server.go`** — Single-file server setup (~900 lines). All route registration, handler wiring, and middleware stack. Services are constructed first, then handlers, then routes. Features are conditionally initialized (fail-closed: missing config disables the feature, doesn't crash).
- **`api/handlers/`** — HTTP handlers. Pattern: struct with service dependencies, methods return `http.HandlerFunc`.
- **`chat/`** — Dual chat orchestrator:
  - `orchestrator.go` — Regex-based intent classification (fallback, always available)
  - `claude_orchestrator.go` — Claude Opus with tool use (primary when `ANTHROPIC_API_KEY` set)
  - `audit_wal.go` — Write-Ahead Log for audit trail
  - `circuit_breaker.go` / `dlq.go` — Failure handling with dead letter queue
  - Commands return `(text, *Artifact, error)` for rich UI rendering
- **`physics/`** — CPM scheduler (forward/backward pass), DHSM duration calculator, SWIM weather model. **Deterministic** — changes require golden master test updates (`cpm_determinism_test.go`).
- **`agents/`** — Autonomous agents (DailyFocus, Procurement, SubLiaison). Each has a standard and Claude-powered variant.
- **`futureshade/`** — Intelligence layer: Tribunal decision engine, shadow execution tracking.
- **`service/`** — Business logic. Services use `pgxpool.Pool` directly with raw SQL. Interfaces in `internal/service/interfaces.go`.
- **`models/`** — Domain models (Project, Task, WBS, Financial, Communication, FeedCard)
- **`worker/`** — Asynq job handlers. Task payloads in `payloads.go`, cron schedules in `scheduler.go`. Runs as separate binary from API server.
- **`middleware/`** — Auth (Clerk JWT via JWKS), role/permission checks, rate limiting, dev auth bypass
- **`config/`** — Environment-based config with `godotenv`. See `.env.example` for all variables.
- **`readiness/`** — Per-service health probes (DB, Clerk, Redis, Vertex, S3, notification providers)

### Conditional Feature Registration

Handlers are only registered in `server.go` if their dependencies are configured:
- `OnboardingHandler` / `VisionHandler` — require `aiClient != nil`
- `StreamChatHandler` — requires `AnthropicAPIKey != ""`
- `GitHubWebhookHandler` — requires `GitHubWebhookSecret != ""`

This means missing an API key won't crash the server — the feature is simply disabled.

### Frontend (`frontend/src/`)

**Component hierarchy:** All components extend `FBElement` (shared styles, `emit()` helper). View/page components extend `FBViewElement` (viewport containment, `onViewActive()` lifecycle hook). Views manage their own scroll — the window NEVER scrolls.

- **`components/`** — Lit web components organized by domain (layout, chat, artifacts, agent, views, settings, shadow, portal)
- **`services/`** — API client, WebSocket/SSE handlers
- **`store/`** — Signals-based reactive state (`@lit-labs/preact-signals`)
- **`types/`** — TypeScript interfaces matching Go `pkg/types/` (Rosetta Stone)

**Path alias:** `@/*` maps to `./src/*` (configured in tsconfig.json). Import as `import { X } from '@/components/base/FBElement'`.

**Component naming:** Tag names use `fb-` prefix (e.g., `fb-my-component`). Files match class names (e.g., `FBElement.ts`).

**Lit decorators:** `experimentalDecorators: true` and `useDefineForClassFields: false` are required for Lit's `@property` decorator.

### AI Integration (`pkg/ai/`)

Vendor-agnostic abstraction with `Client` interface supporting `GenerateContent()` and `GenerateEmbedding()`.

- **`vertex.go`** — Vertex AI (Gemini) for vision, embeddings, general generation
- **`anthropic.go`** — Claude Opus for chat orchestration, reasoning, tool use
- **`factory.go`** — `NewFactory(vertexProjectID, vertexLocation, anthropicKey)` creates clients by provider
- **`types.go`** — Core types: `GenerateRequest`, `GenerateResponse`, `ContentPart` (text, images, tool use/results)

Both providers run simultaneously — Vertex for vision/embeddings, Claude for chat/reasoning.

### Rosetta Stone Type System

Go types in `pkg/types/` ↔ TypeScript types in `frontend/src/types/` must stay in sync.

**Contract test pipeline:** `make contract-test` runs:
1. Go test generates JSON samples from Go structs → `internal/contract_validation/samples/*.json`
2. Frontend validates TS types can parse those JSON samples (via AJV in `frontend/scripts/validate-contract.js`)

Verified types: Forecast, Contact, InvoiceExtraction, GanttData. Adding new shared types requires updates in BOTH Go and TS.

### Worker / Async Jobs

Asynq (Redis-backed) with typed task payloads in `internal/worker/payloads.go`. Runs as a **separate binary** (`cmd/worker/main.go`), not inside the API server.

Key cron schedules: 05:00 UTC procurement, 06:00 UTC daily briefing, 07:00 UTC drift detection, 23:00 UTC expire actions.

Patterns: idempotency checks before processing, circuit breakers for optional features, notification dampening (72h throttle).

### Auth Model

Three auth systems:
- **Clerk** — Main app users. JWT validated via JWKS. Middleware: `RequireAuth`, `RequireRole`, `RequirePermission` (scope-based RBAC)
- **Magic Link** — Portal contacts (field workers). Rate-limited. Separate endpoints under `/api/v1/portal/auth/`
- **API Key** — FB-Brain integration endpoints under `/api/v1/integration/`

**Dev bypass:** `DEV_AUTH_BYPASS=true` injects demo claims without Clerk JWT validation. `CLERK_ISSUER_URL` is required unless this is set. NEVER enable in production.

## Critical Environment Variables

| Variable | Purpose | Required For |
|----------|---------|--------------|
| `DATABASE_URL` | PostgreSQL connection | All |
| `REDIS_URL` | Asynq task queue + sessions | Worker + API |
| `CLERK_ISSUER_URL` | JWKS-based JWT validation | API (unless `DEV_AUTH_BYPASS=true`) |
| `CLERK_SECRET_KEY` | User management / invite flow | Invite flow |
| `ANTHROPIC_API_KEY` | Claude Opus orchestrator | Chat intelligence (optional — falls back to regex) |
| `VERTEX_PROJECT_ID` / `VERTEX_LOCATION` | Gemini API | Document analysis, embeddings (optional) |
| `DEV_AUTH_BYPASS` | Skip JWT validation | Local dev only |

See `.env.example` for the complete list (118 variables with descriptions).

## Git Workflow

- Two persistent branches: `staging` (dev/staging) and `main` (production)
- `staging` → deploys to `staging.futurebuild.ai` via Railway
- `main` → deploys to `project.futurebuild.ai` via Railway (requires PR + 1 review)
- Both branches require all 5 CI checks to pass: Lint, TypeScript Check, Go Tests, Contract Tests, Docker Build
- Direct push to `staging` allowed; `main` requires PR with approval
- Commits should reference spec sections where applicable (e.g., `// See DATA_SPINE_SPEC.md Section 3.3`)

### Docker Build

- **API** (`Dockerfile`): 3-stage (Node frontend builder → Go backend builder → Alpine runtime). Includes `golang-migrate` binary; `entrypoint.sh` runs migrations before startup. Requires `VITE_CLERK_PUBLISHABLE_KEY` build arg.
- **Worker** (`Dockerfile.worker`): 2-stage Go-only build. No frontend assets, no migrations.

## Key Specifications

| Document | Purpose |
|----------|---------|
| `specs/BACKEND_SCOPE.md` | Backend architecture, WBS scope, technology decisions |
| `specs/FRONTEND_SCOPE.md` | Frontend architecture, 3-panel layout, component design |
| `specs/DATA_SPINE_SPEC.md` | Database schema, domain definitions |
| `specs/API_AND_TYPES_SPEC.md` | Shared type definitions (source of truth for Rosetta Stone) |
| `specs/CPM_RES_MODEL_SPEC.md` | CPM scheduler specification |
| `specs/GABLE_LBM_DESIGN_SYSTEM.md` | Design tokens, glassmorphism, typography (GableLBM Industrial Dark) |
| `planning/archive/v1_post_permit/ROADMAP.md` | Beta launch roadmap with phase status |

## Prism Protocol (Dual-Engine Workflow)

The repo uses a two-agent architecture for development:
- **Antigravity (The Brain):** Planning, specs, QA, browser testing — via `.agent/workflows/`
- **Claude Code (The Hands):** File editing, compilation, local testing

Skills in `.agent/skills/`: `product_owner`, `system_architect`, `software_tester`, `l7_gatekeeper`

Workflows in `.agent/workflows/`: `prism.md` (main loop), `devteam.md`, `product.md`, `ops.md`, `deploy.md`, `develop-sprint.md`

## Current State

Phases 10–15 of the Beta Launch Roadmap are complete. Migrations are at `000082`. CI/CD runs on Railway with multi-target Dockerfiles (`api` + `worker`). See `planning/archive/v1_post_permit/ROADMAP.md` for detailed status.
