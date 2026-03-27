# Sprint 2.3: The Physics Trigger

> **Epic:** 2 — The Interrogator Gate (Onboarding Intelligence)
> **Depends On:** Sprint 2.2 (Interrogator Interface)
> **Objective:** Gate schedule generation behind Interrogator approval. CPM engine runs only after the gate opens.

---

## Sprint Tasks

### Task 2.3.1: Implement the Interrogator Gate

**Status:** ✅ Complete

**Concept:** The "Generate Schedule" button is disabled until `InterrogatorAgent` returns `status: SATISFIED`. This creates the "Expert Partner" feel — the AI validates that it has enough information before committing.

**Current State:**
- [onboarding-store.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/store/onboarding-store.ts): `isReadyToCreate` computed signal exists (checks name, address, start_date, square_footage)
- Backend `interrogator_service.md`: `checkReadyToCreate()` returns true when name + address are present
- Gap: No `SATISFIED` status concept — current logic is basic field presence checks

**Atomic Steps:**

1. **Define `InterrogatorStatus` type:**
   ```ts
   // onboarding-store.ts
   export type InterrogatorStatus = 'gathering' | 'clarifying' | 'satisfied' | 'error';
   export const interrogatorStatus = signal<InterrogatorStatus>('gathering');
   ```

2. **Backend: Add `status` field to onboarding response:**
   ```go
   type OnboardingResponse struct {
       // ...existing fields...
       Status string `json:"status"` // "gathering", "clarifying", "satisfied"
   }
   ```

3. **Frontend: Update gate logic:**
   - Replace `isReadyToCreate` with `interrogatorStatus.value === 'satisfied'`
   - Disable "Generate Schedule" button until satisfied
   - Show progress indicator: "AI is validating your project data..."

4. **UX: Visual gate states:**
   - `gathering` → Button hidden, progress shows "Collecting information..."
   - `clarifying` → Button disabled, pulsing amber: "AI has follow-up questions"
   - `satisfied` → Button enabled, green glow: "Ready to Generate Schedule"
   - `error` → Button disabled, red: "Unable to validate — try again"

---

### Task 2.3.2: Connect ScheduleService to CPM Engine

**Status:** ✅ Complete

**Current State:**
- Backend: [cpm.md](file:///home/colton/Desktop/FutureBuild_HQ/XUI/backend/shadow/internal/physics/cpm.md) — placeholder stub, no Go implementation
- Frontend: [fb-view-schedule.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/components/views/fb-view-schedule.ts) — schedule view exists
- Frontend: [mock-schedule-service.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/services/mock-schedule-service.ts) — mock schedule data

**Required Implementation:**

1. **Backend: Implement `cpm.Calculate()`:**
   ```go
   // backend/internal/physics/cpm.go
   type Task struct {
       ID           string
       Name         string
       Duration     int  // days
       Dependencies []string
       EarlyStart   int  // computed
       EarlyFinish  int  // computed
       LateStart    int  // computed
       LateFinish   int  // computed
       TotalFloat   int  // computed
       IsCritical   bool // computed
   }
   
   func Calculate(tasks []Task) (*Schedule, error)
   ```

2. **Backend: Create schedule generation endpoint:**
   - `POST /api/v1/projects/:id/schedule/generate`
   - Accepts WBS from Interrogator extraction
   - Calls `cpm.Calculate()`, stores result
   - Returns `Schedule` with critical path highlighted

3. **Frontend: Wire gate → generation:**
   - On "Generate Schedule" click (only when gate is open)
   - Call schedule generation API
   - Navigate to `/project/:id/schedule` on success
   - Show loading animation during generation

4. **Frontend: Replace mock schedule service:**
   - Wire `fb-view-schedule.ts` to real API
   - Ensure critical path tasks are visually distinguished (red/bold)

---

## Codebase References

| File | Path | Status | Notes |
|------|------|--------|-------|
| onboarding-store.ts | `frontend/src/store/onboarding-store.ts` | Modify | Add InterrogatorStatus signal |
| cpm.md | `backend/shadow/internal/physics/cpm.md` | Stub | Needs Go implementation |
| fb-view-schedule.ts | `frontend/src/components/views/fb-view-schedule.ts` | Existing | Wire to real API |
| mock-schedule-service.ts | `frontend/src/services/mock-schedule-service.ts` | Delete | Replace with real service |

## Verification Plan

- **Manual:** Start onboarding → verify "Generate Schedule" is disabled until AI returns `satisfied`
- **Manual:** Complete interrogation → verify button becomes enabled with green glow
- **Manual:** Click "Generate Schedule" → verify CPM runs and schedule renders
- **Manual:** Verify critical path tasks are visually highlighted
- **Automated:** Unit test `cpm.Calculate()` with known task graph, verify critical path correctness
