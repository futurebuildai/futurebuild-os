# audit/wal — Write-Ahead Log

## Intent
*   **High Level:** Persist user corrections to AI-extracted fields for future model refinement.
*   **Business Value:** Enables data-driven improvement of VisionAgent extraction accuracy. "Users frequently correct 'unit_price_cents' on electrical invoices" → improved prompts.

## Responsibility
*   Provides an append-only Write-Ahead Log for correction events.
*   Persists `CorrectionEntry` records to the `correction_log` PostgreSQL table.
*   Designed for high write throughput with minimal read requirements (batch analytics only).

## CorrectionEntry Struct

```go
package audit

import "time"

// CorrectionEntry represents a single user correction to an AI-extracted field.
type CorrectionEntry struct {
    ID                 string    `json:"id" db:"id"`                    // UUID, auto-generated
    ArtifactID         string    `json:"artifact_id" db:"artifact_id"`
    ArtifactType       string    `json:"artifact_type" db:"artifact_type"` // "invoice"|"budget"|"schedule"
    FieldPath          string    `json:"field_path" db:"field_path"`       // e.g. "line_items[0].unit_price_cents"
    OldValue           any       `json:"old_value" db:"old_value"`         // JSONB
    NewValue           any       `json:"new_value" db:"new_value"`         // JSONB
    OriginalConfidence float64   `json:"original_confidence" db:"original_confidence"`
    UserID             string    `json:"user_id" db:"user_id"`
    Timestamp          time.Time `json:"timestamp" db:"timestamp"`
    CreatedAt          time.Time `json:"created_at" db:"created_at"`       // Server-side insertion time
}
```

## AuditWAL Struct

```go
// AuditWAL is the Write-Ahead Log for correction events.
type AuditWAL struct {
    db *sql.DB
}

// NewAuditWAL creates a new AuditWAL backed by the given database.
func NewAuditWAL(db *sql.DB) *AuditWAL {
    return &AuditWAL{db: db}
}

// Append inserts a CorrectionEntry into the correction_log table.
func (w *AuditWAL) Append(entry CorrectionEntry) error {
    // INSERT INTO correction_log (...) VALUES (...)
}

// AppendBatch inserts multiple CorrectionEntry records in a single transaction.
func (w *AuditWAL) AppendBatch(entries []CorrectionEntry) error {
    // BEGIN; INSERT ...; INSERT ...; COMMIT;
}
```

## PostgreSQL Schema

```sql
CREATE TABLE IF NOT EXISTS correction_log (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    artifact_id         UUID NOT NULL,
    artifact_type       VARCHAR(20) NOT NULL CHECK (artifact_type IN ('invoice', 'budget', 'schedule')),
    field_path          TEXT NOT NULL,
    old_value           JSONB,
    new_value           JSONB,
    original_confidence DECIMAL(3,2) NOT NULL CHECK (original_confidence >= 0 AND original_confidence <= 1),
    user_id             UUID NOT NULL,
    timestamp           TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for analytics queries: "what fields are most frequently corrected?"
CREATE INDEX idx_correction_log_artifact ON correction_log (artifact_type, field_path);
CREATE INDEX idx_correction_log_created ON correction_log (created_at);
```

## Future Use
*   **VisionAgent Prompt Refinement:** Query `correction_log` grouped by `(artifact_type, field_path)` to identify systematic extraction errors.
*   **Confidence Recalibration:** Use correction rate per field to adjust confidence scores in real-time.
*   **Audit Trail:** Provides a complete history of all user corrections for compliance.

## Dependencies
*   **Upstream:** `correction_handler.go` → `HandleCorrectionBatch()`
*   **Downstream:** PostgreSQL `correction_log` table
