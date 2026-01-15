# Handoff: Phase 6 Step 47

**Previous Step:** 46 (Procurement Agent) - **COMPLETED** ✅
**Current Step:** 47 (Sub Liaison Agent)

## Status
- **Worker Active**: `cmd/worker` runs background jobs (Daily Briefing @ 6:00 AM, Procurement Check @ 5:00 AM).
- **Procurement Agent**: Implemented with auto-hydration, weather buffers, notification dampening.
- **Infrastructure**: Redis, Asynq, PostgreSQL all wired.

## Context for Step 47
Implement the **Subcontractor Liaison Agent**. This agent coordinates with trade partners for:
1.  **Start Confirmation**: Confirm site arrival for upcoming tasks.
2.  **Status Checks**: Virtual PM status polling.
3.  **Photo Collection**: Request verification photos.

## Key Dependencies
- `DirectoryService` (Step 38) for contact resolution.
- `NotificationService` for SMS/Email delivery.

## Objectives
1.  Create `internal/agents/sub_liaison.go`.
2.  Implement `GetContactForPhase` integration.
3.  Register worker task if needed (or trigger on-demand).

## Spec References
-   `BACKEND_SCOPE.md` Section 3.5 (Action Engine - Subcontractor Liaison).
-   `PRODUCTION_PLAN.md` Step 47.