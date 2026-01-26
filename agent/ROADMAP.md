# Roadmap

## Phase 7: Frontend - Lit + TypeScript

> **Architecture Pivot**: 3-panel "Agent Command Center" (see FRONTEND_SCOPE.md 3.3)

- [x] Step 50: Initialize Vite project with Lit + TS
- [x] Step 51.1: Frontend Core Architecture (FBElement & Styles)
- [x] Step 51.2: Reactive State Engine (Signals Store)
- [x] Step 51.3: **3-Panel Shell** (Left/Center/Right panels)
- [x] Step 52: **Conversation UI Components** (Message List, Action Cards)
- [x] Step 53: **Agent Activity Log** (Real-time status)
- [x] Step 54: **Mobile Responsive Behavior** (Overlay/Collapse)
- [x] Step 55: **Artifact Panel Renderers** (Gantt, Budget, Invoice)
- [x] Step 56: **Drag-and-Drop Ingestion** (Global overlay, Flicker prevention)
- [x] Step 57: **Real-time WebSocket/SSE Messaging** (RealtimeService, DevTools hooks, Typing indicator)
- [x] Step 58: **Artifact Fixture Testing** (Pure components, shared helpers, fixture harness)
- [x] Step 59: **E2E Demo Readiness** (Full flow verification, polish)

## Phase 8: Production Readiness

- [x] Step 60.1: **Strict Mode & Type Hygiene** (displayTime pre-calculation)
- [x] Step 60.2.1: **Virtualization Infrastructure** (@lit-labs/virtualizer in fb-message-list)
- [x] Step 60.2.2: **Load Test Harness** (LoadTestService, debug buttons)
- [x] Step 60.2.3: **Performance Tuning** (flow layout, verified 60fps)
- [x] Step 61.1: **Security Audit & Hardening** (BOLA fix, Confused Deputy fix, SQL hygiene)
- [x] Step 61.2: **Go Service Mocking & Decoupling** (Interfaces, Mocks, Handler Refactoring, Proof-of-Value Tests)
- [x] Step 62.1: **Infrastructure Setup** (testcontainers, factory, migrations)
- [x] Step 62.2: **Core Integration Tests** (Project Lifecycle Happy Path)
- [x] Step 62.3: **Async Logic Verification** (Task hydration via Asynq)
- [x] Step 62.4: **The "Golden Thread" E2E** (API -> DB -> Worker -> API integration)
- [x] Step 63: **Shadow Site & Protocol** (Internal documentation portal, Dual-Write enforcement)

## Phase 9: FutureShade - The Intelligence Layer

- [x] Step 64: **Initialize FutureShade Service** (internal/futureshade, Event Bus integration)
- [x] Step 65: **Implement "The Tribunal" Interfaces** (Multi-Model Client, Consensus Logic) Completed: 2026-01-25
- [ ] Step 66: **Build The Shadow Viewer** (ShadowDocs + Tribunal Decision Logs UI)
- [ ] Step 67: **Integrate Antigravity Skills** (Agent code execution)
- [ ] Step 68: **Implement Automated PR Review** (GitHub Webhook -> Tribunal -> PR Comment)
- [ ] Step 69: **The "Tree Planting" Ceremony** (Full autonomous bug fix integration test)