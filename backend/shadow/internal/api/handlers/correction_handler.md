# correction_handler

## Intent
*   **High Level:** Accept user corrections to AI-extracted artifact fields for model improvement.
*   **Business Value:** Enables a feedback loop from user corrections back to the VisionAgent, improving extraction accuracy over time.

## Responsibility
*   Receives batch correction events from the frontend via `POST /api/v1/corrections`.
*   Validates the request payload (well-formed events, valid artifact types).
*   Passes validated events to `AuditWAL.Append()` for persistence.
*   Returns `204 No Content` on success, `400 Bad Request` on validation errors.

## API Contract

### `POST /api/v1/corrections`

**Request Body:**
```json
{
    "events": [
        {
            "artifactId": "uuid",
            "artifactType": "invoice",
            "fieldPath": "line_items[0].unit_price_cents",
            "oldValue": 15000,
            "newValue": 17500,
            "originalConfidence": 0.72,
            "timestamp": "2026-02-24T21:14:00Z",
            "userId": "current"
        }
    ]
}
```

**Validation Rules:**
- `events` must be a non-empty array (max 100 items)
- `artifactType` must be one of: `invoice`, `budget`, `schedule`
- `fieldPath` must be a non-empty string
- `originalConfidence` must be 0.0–1.0
- `userId` value `"current"` is resolved to the authenticated principal's ID from the JWT

**Response:** `204 No Content` (no body)

**Error Response:** `400 Bad Request`
```json
{
    "error": { "code": 400, "message": "validation error detail" }
}
```

## Key Logic
*   Decodes JSON request body into `[]CorrectionEvent` struct.
*   Iterates events, resolves `userId = "current"` to `principal.ID`.
*   Calls `AuditWAL.Append()` for each event (or batch insert).
*   On WAL write failure, logs error but returns 204 to avoid blocking frontend save flow.

## Go Signature (Target)
```go
// POST /api/v1/corrections
func HandleCorrectionBatch(wal *audit.AuditWAL) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // ...
    }
}
```

## Dependencies
*   **Upstream:** Frontend `api.corrections.submit()` (fire-and-forget POST)
*   **Downstream:** `internal/audit.AuditWAL.Append()` → PostgreSQL `correction_log` table
