# financial

## Intent
*   **High Level:** Data models for financial summary responses. Defines the contract between the backend FinancialService and the frontend `api.financials` client.
*   **Business Value:** Ensures consistent financial data representation across the stack. The `FinancialSummary` struct is the Rosetta Stone between Go and TypeScript.

## Responsibility
*   Define `FinancialSummary` struct (JSON response shape)
*   Define `CategorySummary` struct (per-category breakdown)
*   Define status enum values for category health

## Key Structs

### `FinancialSummary`
```go
type FinancialSummary struct {
    ProjectID   string            `json:"project_id,omitempty"`
    BudgetTotal int64             `json:"budget_total"`
    SpendTotal  int64             `json:"spend_total"`
    Variance    int64             `json:"variance"`
    Categories  []CategorySummary `json:"categories"`
    LastUpdated time.Time         `json:"last_updated"`
}
```

### `CategorySummary`
```go
type CategorySummary struct {
    Name   string `json:"name"`
    Budget int64  `json:"budget"`
    Spend  int64  `json:"spend"`
    Status string `json:"status"` // "on_track" | "at_risk" | "over_budget"
}
```

### Status Derivation Rules
| Condition | Status |
|-----------|--------|
| spend < 75% of budget | `on_track` |
| spend >= 75% and < 100% of budget | `at_risk` |
| spend >= 100% of budget | `over_budget` |

## Dependencies
*   **Upstream:** `FinancialService` (constructs these structs)
*   **Downstream:** `financial_handler` (serializes to JSON for frontend)
