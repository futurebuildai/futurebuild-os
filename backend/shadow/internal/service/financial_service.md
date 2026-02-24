# financial_service

## Intent
*   **High Level:** Business logic layer for computing financial summaries. Derives "Total Spend" from approved invoices and "Budget" from project configuration.
*   **Business Value:** The single source of truth for all financial calculations. Ensures the "Budget vs. Actual" numbers displayed in the frontend are always derived from real transactional data, not hardcoded values.

## Responsibility
*   Calculate spend totals from approved invoices
*   Retrieve budget configuration from project settings
*   Compute variance and category-level status flags
*   Aggregate across multiple projects for global summaries

## Key Logic

### `GetSummary(ctx context.Context, projectID string) (*FinancialSummary, error)`
1. Query budget from `projects.budget` field (or `project_budgets` table if per-category budgets exist)
2. Call `calculateSpend(ctx, projectID)` to get actual spend
3. Compute variance: `budget_total - spend_total`
4. Build category breakdown with status flags:
   - `on_track`: spend < 75% of category budget
   - `at_risk`: spend >= 75% and < 100% of category budget
   - `over_budget`: spend >= 100% of category budget
5. Set `last_updated` to the most recent invoice approval timestamp

### `GetGlobalSummary(ctx context.Context) (*FinancialSummary, error)`
1. List all projects for the authenticated user's org
2. Aggregate budget totals and spend totals across all projects
3. Compute global variance
4. Roll up categories (e.g., "Labor", "Materials", "Subcontractors")

### `calculateSpend(ctx context.Context, projectID string) (int64, error)`
Core query that derives actual spend from the invoices table:
```sql
SELECT SUM(total_cents)
FROM invoices
WHERE project_id = $1
  AND status = 'approved'
```
Returns `0` if no approved invoices exist (not an error).

### Category Spend Breakdown
```sql
SELECT detected_wbs_code, SUM(total_cents) as spend
FROM invoices
WHERE project_id = $1
  AND status = 'approved'
GROUP BY detected_wbs_code
```
Maps WBS codes to category names via project configuration.

## Dependencies
*   **Upstream:** `financial_handler.go` (HTTP layer)
*   **Downstream:** `invoices` table (for spend), `projects` table (for budget), `project_budgets` table (optional, for category budgets)
