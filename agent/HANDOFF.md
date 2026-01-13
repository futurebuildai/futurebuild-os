# Handoff: Phase 6 Step 43

**Previous Step:** 42 (Mock Ingestion Pipeline) - **COMPLETED**
**Current Step:** 43 (Chat Orchestrator & Intent Mapping)

## Status
- **Step 42 Complete**: Created `test/fixtures/perfect_invoice.json` and `test/integration/pipeline_test.go`. Verified `InvoiceService.SaveExtraction` logic in isolation.
- **Ready for Step 43**: The database layer is regression-tested. We can now build the "Brain" (Chat Orchestrator) that will drive these services.

## Context for Step 43
The Goal is to build the `ChatService` and `Orchestrator`. This is the entry point for all user interaction in the "Chat-First" architecture.
The Orchestrator must:
1. Accept a user message.
2. classify intent (initially we can use regex or simple keywords, backing into Gemini later).
3. Route to the correct service (e.g., `PROCESS_INVOICE` -> `InvoiceService`).
4. Return a structured response (See `API_AND_TYPES_SPEC.md` for `DynamicUIArtifact`).

## Key Files
- `internal/chat/orchestrator.go` (NEW)
- `internal/chat/service.go` (NEW)
- `internal/chat/intents.go` (NEW)
- `pkg/types/chat.go` (Shared types if needed, or use `pkg/types`)

## Next Actions
1. Define `ChatService` struct.
2. Implement `HandleMessage(ctx, user, message)`.
3. Create `Intent` enum (`ProcessInvoice`, `GeneralQuery`, etc.).
4. accessible via `POST /api/v1/chat`.
