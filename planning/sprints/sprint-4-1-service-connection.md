# Sprint 4.1: Service Connection (Real-Time Financials)

> **Epic:** 4 — Real-Time Financials (Destubbing)
> **Depends On:** Sprint 3.1 (Invoice diff view operational)
> **Objective:** Replace all mock financial data with the live backend engine. Money is the most critical data point for investors.

---

## Sprint Tasks

### Task 4.1.1: Delete `mock-financial-service.ts`

**Status:** ✅ Complete

**Current State:**
- [mock-financial-service.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/services/mock-financial-service.ts) — provides `FinancialSummary` with static data
- [fb-view-budget.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/components/views/fb-view-budget.ts) (202L) — imports and uses `mockFinancialService.getSummary('p1')`

**Atomic Steps:**

1. Identify all imports of `mock-financial-service` across codebase
2. Create replacement `financial-service.ts` (or add to existing `api.ts`)
3. Delete `mock-financial-service.ts`
4. Verify no broken imports

---

### Task 4.1.2: Wire `fb-view-budget.ts` to Backend Financial Handler

**Status:** ✅ Complete

**Current State:**
- `fb-view-budget.ts` calls `mockFinancialService.getSummary('p1')` (hardcoded project ID)
- Backend: [financial.md](file:///home/colton/Desktop/FutureBuild_HQ/XUI/backend/shadow/internal/models/financial.md) — placeholder stub
- No `financial_handler.go` exists yet

**Required Backend Implementation:**

1. **Create `backend/internal/api/handlers/financial_handler.go`** [NEW]:
   ```go
   // GET /api/v1/projects/:id/financials/summary
   func HandleFinancialSummary(w http.ResponseWriter, r *http.Request) {
       projectID := chi.URLParam(r, "id")
       summary, err := financialService.GetSummary(ctx, projectID)
       // ...
   }
   
   // GET /api/v1/financials/summary (global across all projects)
   func HandleGlobalFinancialSummary(w http.ResponseWriter, r *http.Request) {
       // Aggregate across all user's projects
   }
   ```

2. **Create `backend/internal/service/financial_service.go`** [NEW]:
   ```go
   type FinancialSummary struct {
       BudgetTotal  int64              `json:"budget_total"`
       SpendTotal   int64              `json:"spend_total"`
       Variance     int64              `json:"variance"`
       Categories   []CategorySummary  `json:"categories"`
       LastUpdated  time.Time          `json:"last_updated"`
   }
   
   func (s *FinancialService) GetSummary(ctx context.Context, projectID string) (*FinancialSummary, error)
   ```

**Required Frontend Changes:**

3. **Add API method:**
   ```ts
   // services/api.ts
   financials: {
       getSummary(projectId: string): Promise<FinancialSummary>,
       getGlobalSummary(): Promise<FinancialSummary>,
   }
   ```

4. **Update `fb-view-budget.ts`:**
   - Import from `api` instead of `mockFinancialService`
   - Use `store.contextState$` to determine scope:
     - Global → `api.financials.getGlobalSummary()`
     - Project → `api.financials.getSummary(projectId)`
   - Add error handling, loading states, retry

---

### Task 4.1.3: Dynamic "Budget vs. Actual" from Approved Invoices

**Status:** ✅ Complete

**Concept:** The "Total Spend" number must be derived from the sum of approved invoices in the database, not hardcoded.

**Atomic Steps:**

1. **Backend: Query approved invoices:**
   ```go
   func (s *FinancialService) calculateSpend(ctx context.Context, projectID string) (int64, error) {
       // SELECT SUM(total_cents) FROM invoices 
       // WHERE project_id = ? AND status = 'approved'
   }
   ```

2. **Backend: Budget source:** Pull from `project_budgets` table (or `projects.budget` field)

3. **Frontend: Show variance with color coding:**
   - Positive variance (under budget) → green
   - Negative variance (over budget) → red
   - Already implemented in `fb-view-budget.ts` CSS (`.positive`, `.negative`)

4. **Real-time updates:** When an invoice is approved (EPIC 3), the budget view should reflect the change. Options:
   - Refetch on view activation (`onViewActive()` already calls `_loadData()`)
   - SSE event from `feed-sse.ts` for budget updates

---

## Codebase References

| File | Path | Status | Notes |
|------|------|--------|-------|
| mock-financial-service.ts | `frontend/src/services/mock-financial-service.ts` | DELETE | Replace with real API |
| fb-view-budget.ts | `frontend/src/components/views/fb-view-budget.ts` | 202L | Modify imports & data loading |
| fb-artifact-budget.ts | `frontend/src/components/artifacts/fb-artifact-budget.ts` | Existing | May need similar wiring |
| budget.ts | `frontend/src/fixtures/budget.ts` | Existing | Test data (keep for tests only) |
| financial.md | `backend/shadow/internal/models/financial.md` | Stub | Needs Go implementation |
| financial_handler.go | `backend/internal/api/handlers/` | [NEW] | API endpoints |
| financial_service.go | `backend/internal/service/` | [NEW] | Business logic |

## Verification Plan

- **Manual:** Navigate to `/budget` → verify data loads from real API (not mock)
- **Manual:** Approve an invoice → navigate to budget → verify "Total Spend" increases
- **Manual:** Switch context (Global → Project) → verify budget view shows correct scope data
- **Automated:** API test: `GET /api/v1/projects/:id/financials/summary` → verify JSON shape matches `FinancialSummary`
- **Automated:** API test: Approve an invoice via API, then fetch summary → verify spend increased
