# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

FutureBuild is an AI-powered construction project management platform. It uses the Residential Construction Path Model (CPM-res1.0) to automate scheduling for residential construction projects, starting from Permit Issued (WBS 5.2).

## Technology Stack

| Layer | Technology |
|-------|------------|
| Backend | Go 1.24+, Chi Router, PostgreSQL 15+ (pgvector), Redis (Asynq) |
| Frontend | Lit 3.0, TypeScript 5.0+ (Strict Mode), Vite, Signals (@lit-labs/preact-signals) |
| AI | Google Vertex AI (Gemini 2.5 Flash/Pro) |
| Auth | Magic link email, JWT tokens |

**Hard Constraints:**
- NO React, NO ORMs (use raw SQL/pgx), NO Python logic (Go only)
- Database is the source of truth; agents are stateless calculators
- All TypeScript must compile with `noImplicitAny` enabled

## Build & Development Commands

```bash
# Start full dev environment (Go API + Vite frontend concurrently)
npm run dev

# Backend only
go run ./cmd/api

# Frontend only
npm --prefix frontend run dev

# Audit (lint + type check)
make audit

# Unit tests (excludes integration)
make test

# Integration tests (requires running DB)
make test-integration

# Contract validation (Go/TS type parity)
make contract-test

# Frontend lint/format
npm --prefix frontend run lint
npm --prefix frontend run lint:fix
npm --prefix frontend run format

# Database migrations
make migrate-up
make migrate-down
make migrate-create name=<migration_name>

# Background worker
make run-worker

# Shadow Protocol (documentation enforcement)
npm run shadow:scaffold  # Generate missing shadow docs
npm run shadow:check     # Verify all source files have shadow docs
```

## Architecture

### Backend Structure (`internal/`)

- **`api/handlers/`** - HTTP handlers (Chi router endpoints)
- **`chat/`** - Chat orchestrator, intent classification, message persistence
- **`physics/`** - CPM scheduler (forward/backward pass), DHSM duration calculator, SWIM weather model
- **`agents/`** - Autonomous agents: DailyFocus, Procurement, SubLiaison, InboundProcessor
- **`service/`** - Business logic services (Auth, Schedule, Vision, Weather, etc.)
- **`models/`** - Domain models (Project, Task, WBS, Financial, Communication)
- **`worker/`** - Asynq job handlers and schedulers
- **`middleware/`** - Auth middleware, rate limiting
- **`platform/`** - Cross-cutting concerns (metrics, DB transactions)

### Frontend Structure (`frontend/src/`)

- **`components/`** - Lit components organized by domain:
  - `base/` - FBElement base class, error boundary
  - `layout/` - 3-panel shell (left/center/right panels)
  - `chat/` - Message list, input bar, action cards
  - `artifacts/` - Gantt, Budget, Invoice renderers
  - `agent/` - Agent activity log
  - `views/` - Page-level view components
- **`services/`** - API client, WebSocket/SSE handlers
- **`store/`** - Signals-based reactive state
- **`types/`** - TypeScript interfaces matching Go `pkg/types` (Rosetta Stone)

### Key Architectural Patterns

1. **Rosetta Stone Type System**: Go types in `pkg/types/` and TypeScript types in `frontend/src/types/` must stay in sync. Run `make contract-test` to verify parity.

2. **Chat-First UI**: 3-panel "Agent Command Center" - Left (projects/threads), Center (chat), Right (artifacts). Users interact via conversation; visual artifacts render inline.

3. **Physics Engine**: CPM scheduling with DHSM duration multipliers and SWIM weather adjustments for pre-dry-in phases. Deterministic calculations with golden master tests.

4. **Service Interfaces**: Core services (Weather, Vision, Directory, Notification) are interface-defined in `internal/service/interfaces.go` with mock implementations for testing.

5. **Shadow Protocol**: Every source file must have a corresponding `.md` documentation file in `frontend/shadow/` or `backend/shadow/`. Run `npm run shadow:check` to verify compliance.

## Testing

```bash
# Run specific Go test
go test -v ./internal/chat/... -run TestOrchestrator

# Run frontend type check
npm --prefix frontend run build

# Integration tests use testcontainers-go
go test -v -tags=integration ./test/integration/...
```

Test fixtures are in `test/fixtures/` and `test/testdata/`. Frontend fixtures are in `frontend/src/fixtures/`.

## Git Workflow

- Default branch for development: `build`
- Do NOT push to `main` or `production` without explicit instruction
- Commits should reference spec sections where applicable (e.g., `// See DATA_SPINE_SPEC.md Section 3.3`)

## Key Specifications

| Document | Purpose |
|----------|---------|
| `specs/BACKEND_SCOPE.md` | Backend architecture, WBS scope, technology decisions |
| `specs/FRONTEND_SCOPE.md` | Frontend architecture, 3-panel layout, component design |
| `specs/DATA_SPINE_SPEC.md` | Database schema, domain definitions |
| `specs/API_AND_TYPES_SPEC.md` | Shared type definitions (source of truth for Rosetta Stone) |
| `planning/ROADMAP.md` | 69-step implementation plan with current progress |
| `agent/SYSTEM_PROMPT.md` | Agent behavior and slash commands |

## Prism Protocol Skills

The repo uses specialized AI skills (in `skills/`):
- `/product` - PRD creation and discovery
- `/devteam` - Engineering execution
- `/ops` - Reliability and incident response
- `/software_engineer` - Context prompt generation

## Current State

Phase 8 (Production Readiness) is complete. Step 63 (Shadow Site & Protocol) is complete. Current focus is Phase 9 (FutureShade - The Intelligence Layer). See `planning/ROADMAP.md` for detailed status.
