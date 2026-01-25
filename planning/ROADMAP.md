# FutureBuild: Zero to Production - 69 Step Plan

**Version:** 1.2.0
**Date:** 2026-01-05
**Focus:** Lit/TypeScript Architecture, Rosetta Stone Type System, and Robust Multi-Layer Validation

---

## Overview

This plan outlines 69 sequential steps to take FutureBuild from zero to production. The strategy emphasizes type safety through a shared "Rosetta Stone," a modern reactive frontend using Lit and TypeScript, and automated validation checkpoints at every layer.

---

## Phase 0: Foundation & CI/CD (Steps 1-8)

| Status | Step | Task | Why Upfront | Est. Days |
|--------|------|------|-------------|-----------|
| [x] | 1 | Set up Git branching strategy (main, develop, feature/*) | Everything depends on version control | 0.5 |
| [x] | 2 | Configure GitHub Actions pipeline (linting/tests) | Catch errors early, enforce consistency | 1 |
| [x] | 3 | Set up automated testing pipeline (go test) | Tests run on every PR from day 1 | 1 |
| [x] | 4 | Configure pre-commit hooks (gofmt, go vet, staticcheck) | Code quality before it hits repo | 0.5 |
| [x] | 5 | Set up staging environment on Digital Ocean | Test deployments safely | 1 |
| [x] | 6 | Configure environment variable management (dev/staging/prod) | Secrets flow correctly from start | 0.5 |
| [x] | 7 | Set up database migrations framework (golang-migrate) | Schema changes tracked from beginning | 1 |
| [x] | 8 | Create Docker Compose for local dev parity | Consistent dev environments | 1 |

---

## Phase 1: Database & Core Models (Steps 9-16)

| Status | Step | Task | Dependencies | Est. Days |
|--------|------|------|--------------|-----------|
| [x] | 9 | Provision PostgreSQL 15 with pgvector extension | Step 5 | 0.5 |
| [x] | 10 | Create Identity & Access models (Org, User, Project) | Step 7 | 2 |
| [x] | 11 | Implement WBS_TEMPLATE and WBS_TASK models | Step 10 | 1 |
| [x] | 12 | **Create Domain 3 Financial tables** (PROJECT_BUDGETS, INVOICES) | Step 11 | 1.5 |
| [x] | 13 | **Create Domain 4 Communication tables** (CONTACTS, ASSIGNMENTS, LOGS) | Step 11 | 1.5 |
| [x] | 14 | Create PROJECT_TASK model with dependency fields | Step 11 | 1 |
| [x] | 15 | **Seed data for WBS_TASKS including "Ghost Predecessors" (WBS 6.x)** | Step 11 | 1 |
| [x] | 16.1 | Implement DB indexing and health check endpoint | Step 9 | 0.5 |
| [x] | 16.2 | **Implement Project CRUD APIs** (POST/GET) with Multi-Tenancy gates | Step 16.1 | 0.5 |

---

## Phase 2: The Rosetta Stone - Type System (Steps 17-20)

| Status | Step | Task | Dependencies | Est. Days |
|--------|------|------|--------------|-----------|
| [x] | 17 | **Create `pkg/types` in Go**: Define Enums and Structs from API_AND_TYPES_SPEC.md | Step 11 | 1 |
| [x] | 18 | **Create `frontend/src/types`**: Define matching TypeScript Interfaces and Enums | Step 17 | 1 |
| [x] | 19 | **Define Go Service Interfaces**: Weather, Vision, Notification, Directory | Step 17 | 1 |
| [x] | 20 | **Contract Validation Test**: Write a test suite that marshals all Go structs to JSON and validates them against the TypeScript definitions | Step 18 | 1 |

---

## Phase 3: Authentication & Authorization (Steps 21-25)

| Status | Step | Task | Dependencies | Est. Days |
|--------|------|------|--------------|-----------|
| [x] | 21 | Implement magic link email authentication | Step 10, 17 | 2 |
| [x] | 22 | Create JWT token generation and validation logic | Step 21 | 1 |
| [x] | 23 | Build role-based permission middleware (Admin, Builder, Client, Sub) | Step 22 | 1.5 |
| [x] | 24 | Create portal access tokens for external users | Step 23 | 1 |
| [x] | 25 | Add rate limiting for auth endpoints | Step 21 | 0.5 |

---

## Phase 4: Physics Engine - Core Scheduling (Steps 26-34)

| Status | Step | Task | Dependencies | Est. Days |
|--------|------|------|--------------|-----------|
| [x] | 26 | Implement DHSM calculator with multiplier logic | Step 14, 17 | 2 |
| [x] | 27 | Build dependency graph using gonum/graph | Step 14 | 2 |
| [x] | 28 | Implement CPM forward pass (ES, EF) | Step 27 | 1 |
| [x] | 29 | Implement CPM backward pass (LS, LF, float, critical path) | Step 28 | 1 |
| [x] | 30 | Implement "Event Duration Locking" (fixed durations for Inspections) | Step 26 | 0.5 |
| [x] | 31 | Add SWIM weather overlay for pre-dry-in tasks | Step 26 | 2 |
| [x] | 32 | Build schedule recalculation trigger system | Step 29 | 1 |
| [x] | 33 | **Golden Master Physics Test**: Create a test suite that runs the DHSM and CPM calculators against a dataset of 50 pre-calculated scenarios | Step 32 | 1 |
| [x] | 34 | **Cycle Detection Test**: Verify the CPM solver correctly handles and rejects circular dependencies | Step 27 | 0.5 |

---

## Phase 5: Context Engine - AI Integration (Steps 35-42)

| Status | Step | Task | Dependencies | Est. Days |
|--------|------|------|--------------|-----------|
| [x] | 35 | Set up Vertex AI client and PDF upload pipeline (S3) | Step 6, 9 | 2 |
| [x] | 36 | Implement RAG pipeline with pgvector embeddings | Step 35 | 3 |
| [x] | 37 | **Implement Invoice Processor logic**: PDF -> `InvoiceExtraction` JSON | Step 17, 36 | 3 |
| [x] | 38 | **Implement DirectoryService lookup logic** (Project Phase -> Contact) | Step 13, 19 | 1.5 |
| [x] | 39 | Add confidence scoring and human review flags | Step 37 | 1 |
| [x] | 40 | Build site photo verification flow | Step 35 | 2 |
| [x] | 40b | **Upgrade Vertex AI SDK**: Refactor `pkg/ai` to support Gemini 2.5 Flash image payloads | Step 40 | 1 |
| [x] | 41 | Create document re-processing and audit trail system | Step 36 | 1 |
| [x] | 42 | **Mock Ingestion Pipeline**: Create a test fixture that injects "perfect" JSON to verify DB logic | Step 37, 12 | 1 |

---

## Phase 6: Action Engine - Chat & Agents (Steps 43-49)

| Status | Step | Task | Dependencies | Est. Days |
|--------|------|------|--------------|-----------|
| [x] | 43.1 | **Domain Modeling (Types)**: Define strict data contracts for Chat domain | Step 36, 37 | 1 |
| [x] | 43.2 | **Intent Classification (Router)**: Implement KeywordClassifier (V1 MVP) | Step 43.1 | 1 |
| [x] | 43.3 | **Orchestration Service (Executor)**: Build traffic controller and logic flow | Step 43.2 | 1 |
| [x] | 43.4 | **API Handler (Interface)**: Expose orchestrator via HTTP with strict security | Step 43.3 | 1 |
| [x] | 43.5 | **Wiring & Assembly**: Register components and apply AuthMiddleware | Step 43.4 | 0.5 |
| [x] | 43.6 | **Verification**: Verify endpoint with mock Auth Token and DB check | Step 43.5 | 0.5 |
| [x] | 44 | Implement internal artifact mapping (Tool Output -> ArtifactType) | Step 43 | 1 |
| [x] | 45 | Create prioritized daily briefing job (Asynq) | Step 29, 43 | 2 |
| [x] | 46 | **Update Procurement Agent**: Lead-times + Weather/Buffer calculations | Step 26, 31, 43 | 2 |
| [x] | 47 | **Update Sub Liaison Agent**: Use DirectoryService for contact resolving | Step 19, 38 | 1.5 |
| [x] | 48 | Implement inbound message processing and state-machine updates | Step 43, 47 | 2 |
| [x] | 49 | **Time-Travel Agent Simulation**: Create an integration test with mocked Clock interface | Step 48 | 2 |

---

## Phase 7: Frontend - Lit + TypeScript (Steps 50-59)

> [!IMPORTANT]
> **UI Architecture Pivot (v1.3.0)**: 3-panel "Agent Command Center" layout replaces SaaS dashboard.
> See FRONTEND_SCOPE.md Section 3.3 for details.

| Status | Step | Task | Dependencies | Est. Days |
|--------|------|------|--------------|-----------|
| [x] | 50 | **Initialize Vite project with Lit + TS**; Configure `@types` aliases | Step 18 | 1 |
| [x] | 51.1 | **Base Architecture**: `FBElement`, global styles, and registry | Step 50 | 1 |
| [x] | 51.2 | **Reactive State Engine**: Signals Store (`store.ts`) & Service Layer | Step 51.1 | 1 |
| [x] | 51.3 | **3-Panel Shell**: Left (Projects/Threads), Center (Chat), Right (Artifacts) | Step 51.2 | 2 |
| [x] | 52 | **Conversation UI Components**: Message List, Action Cards, Input Bar | Step 51.3 | 2 |
| [x] | 53 | **Agent Activity Log**: Real-time status w/ expanding details | Step 52 | 1 |
| [x] | 54 | **Mobile Responsive Behavior**: Panel overlays & collapse logic | Step 53 | 2 |
| [x] | 55 | **Artifact Panel Renderers**: Gantt, Budget, Invoice components | Step 54 | 3 |
| [x] | 56 | **Drag-and-Drop Ingestion**: Invoice upload zone in specialized input | Step 55 | 1.5 |
| [x] | 57 | **Real-time Messaging**: WebSocket/SSE wiring | Step 56 | 2 |
| [x] | 58 | **Artifact Fixture Testing**: Wire components to Store data & Fixtures | Step 55 | 1 |
| [x] | 59 | **E2E Demo Readiness**: Full flow verification, accessibility, polish | Step 58 | 2 |

---

## Phase 8: Production Readiness (Steps 60-63)

| Status | Step | Task | Dependencies | Est. Days |
|--------|------|------|--------------|-----------|
| [x] | 60.1 | **Strict Mode & Type Hygiene**: Enforce `no-explicit-any` | Step 59 | 1 |
| [x] | 60.2.1 | **Virtualization Infrastructure**: Implement `@lit-labs/virtualizer` in `fb-message-list` | Step 60.1 | 1 |
| [x] | 60.2.2 | **Load Test Harness**: Create `LoadTestService` to pump 1,000+ messages | Step 60.2.1 | 0.5 |
| [x] | 60.2.3 | **Performance Tuning**: Profile and tune `overscan` buffers for 60fps scrolling | Step 60.2.2 | 0.5 |
| [x] | 61.1 | **Security Audit & Hardening**: Run static analysis, audit SQL queries, verify RBAC | All previous | 1.5 |
| [x] | 61.2 | **Go Service Mocking**: Implement "Spy" and "Stub" mocks for Weather, Vision, Directory | 61.1 | 1.5 |
| [x] | 62.1 | **Integration Test Infrastructure**: Set up `TestMain` with `testcontainers-go` | Step 61.2 | 1 |
| [x] | 62.2 | **Core API Integration Suite**: Write `handler_test.go` suites for Projects and Tasks | Step 62.1 | 1 |
| [x] | 62.3 | **Agent Loop Verification**: Integration tests for Asynq workers | Step 62.2 | 1.5 |
| [x] | 62.4 | **The "Golden Thread" E2E**: A single "Life of a Project" end-to-end test | Step 62.3 | 1.5 |
| [x] | 63 | **Implement Shadow Site & Protocol**: Deploy internal documentation portal | All previous | 3 |

## Phase 9: FutureShade - The Intelligence Layer (Steps 64-69)

| Status | Step | Task | Dependencies | Est. Days |
|--------|------|------|--------------|-----------|
| [x] | 64 | **Initialize FutureShade Service**: Create `internal/futureshade` and `frontend/futureshade` | Step 63 | 2 |
| [ ] | 65 | **Implement "The Tribunal" Interfaces**: Define Multi-Model Client and Consensus Logic | Step 64 | 3 |
| [ ] | 66 | **Build The Shadow Viewer**: Specialized UI for ShadowDocs and Tribunal Decision Logs | Step 50, 64 | 3 |
| [ ] | 67 | **Integrate Antigravity Skills**: Allow FutureShade to call `internal/agents` | Step 49, 65 | 2 |
| [ ] | 68 | **Implement Automated PR Review**: GitHub Webhook -> Tribunal -> PR Comment | Step 67 | 2 |
| [ ] | 69 | **The "Tree Planting" Ceremony**: Final integration test where FutureShade auto-diagnoses and fixes | Step 68 | 2 |

---

## Definition of Done

1. Code passes all Go linting (golangci-lint).
2. **TypeScript compiles with no 'any' types (Strict Mode).**
3. Unit tests written and passing (>80% coverage).
4. **Go Service Interfaces (Weather, Vision, etc.) mocked and tested.**
5. Database migrations tested (forward and rollback).
6. Deployed to staging and manually verified.
7. Documentation updated (API and README).
8. **All Phase Checkpoints (Contract, Physics, Time-Travel) must pass.**
9. **No logic committed without a corresponding Replay Test.**

---

*Document Version: 1.2.0*
