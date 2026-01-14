# Handoff: Phase 6 Step 43.5

**Previous Step:** 43.4 (API Handler) - **COMPLETED** ✅
**Current Step:** 43.5 (Wiring & Assembly)

## Status
- **Step 43.4 Complete**: `ChatHandler` implemented with strict security (Context-based identity), input validation, logs, and 100% test coverage.
- **Components Ready**:
    - `KeywordClassifier` (Step 43.2)
    - `Orchestrator` (Step 43.3)
    - `ChatHandler` (Step 43.4)
- **Ready for Step 43.5**: Wire it all together in `internal/server/server.go`.

## Context for Step 43.5
We have all the Lego blocks (`Orchestrator`, `Services`, `Handler`). Now we need to snap them onto the baseplate (`Server`).

1.  **Dependency Injection**: The `Orchestrator` needs access to `TaskService`, `ScheduleService`, etc. (Many of these might still be mocks or stubs if not fully implemented in previous phases, but we must wire what we have or defined mocks).
2.  **Route Registration**: Map `POST /api/v1/chat` to our new handler.
3.  **Middleware**: Ensure the `AuthMiddleware` is wrapping this route.

## Requirements
1.  Verify `internal/server/server.go` (or `routes.go`) structure.
2.   Instantiate `chat.NewOrchestrator(...)` with required dependencies.
3.  Instantiate `handlers.NewChatHandler(...)`.
4.  Register Route: `r.Post("/api/v1/chat", chatHandler.HandleChat)`.

## Key Files
- `internal/server/server.go`
- `cmd/server/main.go`
- `internal/api/handlers/chat_handler.go`

## Spec References
- `PRODUCTION_PLAN.md` Step 43.5
