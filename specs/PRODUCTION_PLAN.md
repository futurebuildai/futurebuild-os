# FutureBuild: Zero to Production - 59 Step Plan

**Version:** 1.2.0  
**Date:** 2026-01-05  
**Focus:** Lit/TypeScript Architecture, Rosetta Stone Type System, and Robust Multi-Layer Validation

---

## Overview

This plan outlines 59 sequential steps to take FutureBuild from zero to production. The strategy emphasizes type safety through a shared "Rosetta Stone," a modern reactive frontend using Lit and TypeScript, and automated validation checkpoints at every layer.

---

## Phase 0: Foundation & CI/CD (Steps 1-8)

| Step | Task | Why Upfront | Est. Days |
|------|------|-------------|-----------|
| 1 | Set up Git branching strategy (main, develop, feature/*) | Everything depends on version control | 0.5 |
| 2 | Configure GitHub Actions pipeline (linting/tests) | Catch errors early, enforce consistency | 1 |
| 3 | Set up automated testing pipeline (go test) | Tests run on every PR from day 1 | 1 |
| 4 | Configure pre-commit hooks (gofmt, go vet, staticcheck) | Code quality before it hits repo | 0.5 |
| 5 | Set up staging environment on Digital Ocean | Test deployments safely | 1 |
| 6 | Configure environment variable management (dev/staging/prod) | Secrets flow correctly from start | 0.5 |
| 7 | Set up database migrations framework (golang-migrate) | Schema changes tracked from beginning | 1 |
| 8 | Create Docker Compose for local dev parity | Consistent dev environments | 1 |

---

## Phase 1: Database & Core Models (Steps 9-16)

| Step | Task | Dependencies | Est. Days |
|------|------|--------------|-----------|
| 9 | Provision PostgreSQL 15 with pgvector extension | Step 5 | 0.5 |
| 10 | Create Identity & Access models (Org, User, Project) | Step 7 | 2 |
| 11 | Implement WBS_TEMPLATE and WBS_TASK models | Step 10 | 1 |
| 12 | **Create Domain 3 Financial tables** (PROJECT_BUDGETS, INVOICES) | Step 11 | 1.5 |
| 13 | **Create Domain 4 Communication tables** (CONTACTS, ASSIGNMENTS, LOGS) | Step 11 | 1.5 |
| 14 | Create PROJECT_TASK model with dependency fields | Step 11 | 1 |
| 15 | **Seed data for WBS_TASKS including "Ghost Predecessors" (WBS 6.x)** | Step 11 | 1 |
| 16.1 | Implement DB indexing and health check endpoint | Step 9 | 0.5 |
| 16.2 | **Implement Project CRUD APIs** (POST/GET) with Multi-Tenancy gates | Step 16.1 | 0.5 |

---

## Phase 2: The Rosetta Stone - Type System (Steps 17-20)

| Step | Task | Dependencies | Est. Days |
|------|------|--------------|-----------|
| 17 | **Create `pkg/types` in Go**: Define Enums and Structs from [API_AND_TYPES_SPEC.md](file:///home/colton/Replit%20Specs/API_AND_TYPES_SPEC.md) | Step 11 | 1 |
| 18 | **Create `frontend/src/types`**: Define matching TypeScript Interfaces and Enums | Step 17 | 1 |
| 19 | **Define Go Service Interfaces**: Weather, Vision, Notification, Directory | Step 17 | 1 |
| 20 | **Contract Validation Test**: Write a test suite that marshals all Go structs to JSON and validates them against the TypeScript definitions (or a shared JSON Schema) to ensure strict interoperability. | Step 18 | 1 |

---

## Phase 3: Authentication & Authorization (Steps 21-25)

| Step | Task | Dependencies | Est. Days |
|------|------|--------------|-----------|
| 21 | Implement magic link email authentication | Step 10, 17 | 2 |
| 22 | Create JWT token generation and validation logic | Step 21 | 1 |
| 23 | Build role-based permission middleware (Admin, Builder, Client, Sub) | Step 22 | 1.5 |
| 24 | Create portal access tokens for external users | Step 23 | 1 |
| 25 | Add rate limiting for auth endpoints | Step 21 | 0.5 |

---

## Phase 4: Physics Engine - Core Scheduling (Steps 26-34)

| Step | Task | Dependencies | Est. Days |
|------|------|--------------|-----------|
| 26 | Implement DHSM calculator with multiplier logic | Step 14, 17 | 2 |
| 27 | Build dependency graph using gonum/graph | Step 14 | 2 |
| 28 | Implement CPM forward pass (ES, EF) | Step 27 | 1 |
| 29 | Implement CPM backward pass (LS, LF, float, critical path) | Step 28 | 1 |
| 30 | Implement "Event Duration Locking" (fixed durations for Inspections) | Step 26 | 0.5 |
| 31 | Add SWIM weather overlay for pre-dry-in tasks | Step 26 | 2 |
| 32 | Build schedule recalculation trigger system | Step 29 | 1 |
| 33 | **Golden Master Physics Test**: Create a test suite that runs the DHSM and CPM calculators against a dataset of 50 pre-calculated scenarios (defined in a CSV). Assert that code output matches expected float values to 0.01 precision. | Step 32 | 1 |
| 34 | **Cycle Detection Test**: Verify the CPM solver correctly handles and rejects circular dependencies. | Step 27 | 0.5 |

---

## Phase 5: Context Engine - AI Integration (Steps 35-42)

| Step | Task | Dependencies | Est. Days |
|------|------|--------------|-----------|
| 35 | Set up Vertex AI client and PDF upload pipeline (S3) | Step 6, 9 | 2 |
| 36 | Implement RAG pipeline with pgvector embeddings | Step 35 | 3 |
| 37 | **Implement Invoice Processor logic**: PDF -> `InvoiceExtraction` JSON | Step 17, 36 | 3 |
| 38 | **Implement DirectoryService lookup logic** (Project Phase -> Contact) | Step 13, 19 | 1.5 |
| 39 | Add confidence scoring and human review flags | Step 37 | 1 |
| 40 | Build site photo verification flow | Step 35 | 2 |
| 40b | **Upgrade Vertex AI SDK**: Refactor `pkg/ai` to usage `google.golang.org/genai` to support Gemini 2.5 Flash image payloads. | Step 40 | 1 | [x] |
| 41 | Create document re-processing and audit trail system | Step 36 | 1 | [x] |
| 42 | **Mock Ingestion Pipeline**: Create a test fixture that injects "perfect" JSON (simulating Gemini) into the system to verify that the INVOICES and PROJECT_TASKS tables update correctly, isolating DB logic from AI latency. | Step 37, 12 | 1 | [x] |

---

## Phase 6: Action Engine - Chat & Agents (Steps 43-49)

| Step | Task | Dependencies | Est. Days |
|------|------|--------------|-----------|
| 43.1 | **Domain Modeling (Types)**: Define strict data contracts for Chat domain (Intent, ChatRequest, ChatResponse) in `internal/chat/types.go` | Step 36, 37 | x|
| 43.2 | **Intent Classification (Router)**: Implement KeywordClassifier (V1 MVP) and tests in `internal/chat/intents.go` | Step 43.1 | x|
| 43.3 | **Orchestration Service (Executor)**: Build traffic controller, persistence, and logic flow in `internal/chat/orchestrator.go` | Step 43.2 |x|
| 43.4 | **API Handler (Interface)**: Expose orchestrator via HTTP with strict security in `internal/api/handlers/chat_handler.go` | Step 43.3 |x|
| 43.5 | **Wiring & Assembly**: Register components in `internal/server/server.go` and apply AuthMiddleware | Step 43.4 |x|
| 43.6 | **Verification**: Verify endpoint with mock Auth Token and DB check | Step 43.5 |x|
| 44 | Implement internal artifact mapping (Tool Output -> ArtifactType) | Step 43 | x|
| 45 | Create prioritized daily briefing job (Asynq) | Step 29, 43 | 2 | [x] |
| 46 | **Update Procurement Agent**: Lead-times + Weather/Buffer calculations | Step 26, 31, 43 | 2 |
| 47 | **Update Sub Liaison Agent**: Use DirectoryService for contact resolving | Step 19, 38 | 1.5 |
| 48 | Implement inbound message processing and state-machine updates | Step 43, 47 | 2 |
| 49 | **Time-Travel Agent Simulation**: Create an integration test using a mocked Clock interface. Fast-forward a test project by 30 days, triggering cron jobs at every interval, and assert that the COMMUNICATION_LOGS table contains the expected "Order Material" and "Start Confirmation" messages. | Step 48 | 2 |

---

## Phase 7: Frontend - Lit + TypeScript (Steps 50-56)

| Step | Task | Dependencies | Est. Days |
|------|------|--------------|-----------|
| 50 | **Initialize Vite project with Lit + TS**; Configure `@types` aliases | Step 18 | 1 |
| 51 | **Implement Base Component extensions** and Signals-based Store | Step 50 | 2 |
| 52 | **Build Chat Interface Container with Message List and Ephemeral Cards** | Step 51 | 2 |
| 52.5 | **Implement WebSocket/SSE real-time messaging** for agent responses | Step 52 | 2 |
| 53 | **Build Specialized Artifact components** (Invoice, Budget, Gantt, Rolodex) | Step 18, 44 | 3 |
| 53.5 | **Implement Dynamic Agent UI Renderer** (Recursively render `<fb-dynamic-renderer>` from JSON) | Step 53 | 2 |
| 54 | **Implement Drag-and-Drop zone** for invoice ingestion | Step 35, 51 | 1.5 |
| 55 | Finalize responsive mobile navigation and state hydration | Step 52 | 2 |
| 56 | **Artifact Fixture Testing**: Implement a Storybook-style harness to render <fb-artifact-invoice>, <fb-artifact-budget>, and <fb-artifact-gantt> in isolation with various data states (Loading, Error, Empty, Full) to verify visual stability. | Step 53 | 1 |

---

## Phase 8: Production Readiness (Steps 57-59)

| Step | Task | Dependencies | Est. Days |
|------|------|--------------|-----------|
| 57 | Load testing and **Strict Mode TypeScript validation** | All previous | 2 |
| 58 | Security audit and **Go Interface mock testing** | All previous | 3 |
| 59 | Set up production monitoring and blue-green deployment | Step 5, 8 | 2 |

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
