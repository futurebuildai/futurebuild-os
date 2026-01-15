# Handoff: Phase 6 Step 46

**Previous Step:** 45 (Daily Briefing Job) - **COMPLETED** ✅
**Current Step:** 46 (Update Procurement Agent)

## Status
- **Worker Active**: `cmd/worker` is ready to run background jobs.
- **Daily Briefing**: Scheduled for 6:00 AM, using `DailyFocusAgent`.
- **Infrastructure**: Redis connection and Asynq server are wired.

## Context for Step 46
We need to implement the **Procurement Agent**. This agent monitors long-lead items (Trusses, Windows, HVAC) and calculates:
1.  **Order Date**: `Need Date` (from Schedule) - `Lead Time` - `Buffer`.
2.  **Weather Buffer**: If `SWIM` predicts rain for the rough-in phase, extend the buffer? (Check logic).

## Objectives
1.  Refactor `ProcurementAgent` (if exists) or create `internal/agents/procurement.go`.
2.  Implement `CalculateOrderDate` logic.
3.  Register a new task `task:procurement_check` in the worker scheduler (e.g., run nightly).

## Spec References
-   `BACKEND_SCOPE.md` Section 2.5 (Long-Lead Items) & 3.5 (Action Engine).
-   `PRODUCTION_PLAN.md` Step 46.