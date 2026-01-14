# Handoff: Phase 6 Step 43.3

**Previous Step:** 43.2 (Intent Classification) - **COMPLETED** ✅
**Current Step:** 43.3 (Orchestration Service)

## Status
- **Step 43.2 Complete**: `ClassifyIntent` in `internal/chat/intents.go` with `Keyword` type and deterministic priority.
- **CTO Audit**: APPROVED (2x - Post-Refinement)
- **Ready for Step 43.3**: Orchestration Service.

## Context for Step 43.3
Build the central traffic controller (`Orchestrator`) that:
1. Takes a `ChatRequest`.
2. Classifies it using `ClassifyIntent`.
3. Persists the message to `chat_messages` (See `DATA_SPINE_SPEC.md` Section 5.3).
4. Switches on Intent to execute placeholder logic.
5. Returns a `ChatResponse`.

## Requirements
1. Create `internal/chat/orchestrator.go`.
2. Implement `Orchestrator` struct with DB pool dependency.
3. Implement `ProcessRequest(ctx, userID, req) (*ChatResponse, error)`.
4. Write tests in `internal/chat/orchestrator_test.go`.

## Key Files
- `internal/chat/orchestrator.go` (NEW)
- `internal/chat/orchestrator_test.go` (NEW)
- `internal/chat/intents.go` (Classifier)
- `internal/chat/types.go` (Data Contracts)

## Spec References
- `PRODUCTION_PLAN.md` Step 43.3
- `DATA_SPINE_SPEC.md` Section 5.3 (CHAT_MESSAGES)
- `BACKEND_SCOPE.md` Section 3.5 (Action Engine)
