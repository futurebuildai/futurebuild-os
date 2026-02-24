# financial_handler

## Intent
*   **High Level:** HTTP handler layer for financial summary endpoints. Routes incoming requests to the FinancialService.
*   **Business Value:** Enables the frontend to replace mock financial data with live backend queries, providing real-time "Budget vs. Actual" derived from approved invoices.

## Responsibility
*   Parse project ID from URL parameters
*   Validate request authorization (user must own the project or belong to the org)
*   Delegate to `FinancialService.GetSummary()` or `FinancialService.GetGlobalSummary()`
*   Return JSON `FinancialSummary` response

## Endpoints

### `GET /api/v1/projects/:id/financials/summary`
Returns the financial summary for a specific project:
- Budget total (from `projects.budget` or `project_budgets` table)
- Spend total (SUM of approved invoice `total_cents`)
- Variance (budget - spend)
- Category breakdown with status flags

### `GET /api/v1/financials/summary`
Returns an aggregated financial summary across all projects the user has access to.
Same response shape as the project-scoped endpoint but without `project_id`.

## Response Shape
```json
{
    "project_id": "uuid (optional)",
    "budget_total": 1250000,
    "spend_total": 450000,
    "variance": 800000,
    "last_updated": "2026-02-24T12:00:00Z",
    "categories": [
        {
            "name": "Site Work",
            "budget": 150000,
            "spend": 148000,
            "status": "on_track"
        }
    ]
}
```

## Key Logic
*   Project-scoped: Extract `projectID` from `chi.URLParam(r, "id")`
*   Global: No project filter — aggregate across all org projects
*   Both endpoints require authenticated principal from context

## Dependencies
*   **Upstream:** Frontend `api.financials.getSummary()` / `api.financials.getGlobalSummary()`
*   **Downstream:** `FinancialService.GetSummary()`, `FinancialService.GetGlobalSummary()`
