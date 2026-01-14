# Handoff: Phase 6 Step 43.4

**Previous Step:** 43.3 (Orchestration Service) - **COMPLETED** ✅
**Current Step:** 43.4 (API Handler)

## Status
- **Step 43.3 Complete**: `Orchestrator` service built with dependency injection and full persistence flow.
- **CTO Audit**: APPROVED (Zero-Trust Passed).
- **Ready for Step 43.4**: Expose the service via HTTP.

## Context for Step 43.4
We need to expose the `Orchestrator` to the frontend via a secured HTTP endpoint.
1.  **Handler**: `internal/api/handlers/chat_handler.go`
2.  **Input**: JSON `ChatRequest` (ProjectID, Message).
3.  **Security**: Extract `UserID` from the request context (RBAC Middleware).
4.  **Flow**: `Handler` -> `Orchestrator.ProcessRequest` -> `JSON Response`.

## Requirements
1.  Create `internal/api/handlers/chat_handler.go`.
2.  Define `ChatHandler` struct with `*chat.Orchestrator`.
3.  Implement `HandleChat(w http.ResponseWriter, r *http.Request)`.
4.  **Strict Security**: Ensure `UserID` is strictly retrieved from `r.Context()`. DO NOT trust user input for identity.

## Key Files
- `internal/api/handlers/chat_handler.go` (NEW)
- `internal/chat/orchestrator.go` (Dependency)
- `internal/chat/types.go` (Contracts)

## Spec References
- `PRODUCTION_PLAN.md` Step 43.4
- `BACKEND_SCOPE.md` Section 5.2 (API Structure)
