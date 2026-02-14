# Sprint 3.2: Interactive Learning

> **Epic:** 3 — Intelligent Artifacts (Interactive "Diffs")
> **Depends On:** Sprint 3.1 (Invoice Diff View with confidence data)
> **Objective:** When users correct AI-extracted values, log corrections to the backend for model improvement.

---

## Sprint Tasks

### Task 3.2.1: Send CorrectionEvent on User Edits

**Status:** ⬜ Not Started

**Current State:**
- [fb-artifact-invoice.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/components/artifacts/fb-artifact-invoice.ts) has `_updateItem()`, `_save()` — edit and save workflow already exists
- No correction event dispatch mechanism

**Concept:** When a user changes a value that was AI-extracted, capture the diff (old value → new value) and send it to the backend.

**Atomic Steps:**

1. **Define `CorrectionEvent` type:**
   ```ts
   // types/events.ts [NEW or extend existing]
   export interface CorrectionEvent {
       artifactId: string;
       artifactType: 'invoice' | 'budget' | 'schedule';
       fieldPath: string;    // e.g., "line_items[2].unit_price_cents"
       oldValue: unknown;    // AI-extracted value
       newValue: unknown;    // User-corrected value
       originalConfidence: number;
       timestamp: string;
       userId: string;
   }
   ```

2. **Capture diff in `_save()`:**
   - Before saving, compare `_draftItems` against original `data.extraction_result.items`
   - For each changed field, create a `CorrectionEvent`
   - Only create events for fields that had AI-sourced values (check `fieldConfidences`)

3. **Create API call:**
   ```ts
   // services/api.ts — add to api object
   corrections: {
       submit(events: CorrectionEvent[]): Promise<void>
   }
   ```
   - `POST /api/v1/corrections` — batch submit correction events

4. **Fire-and-forget:** Don't block the save workflow on correction submission. Log errors but don't show to user.

---

### Task 3.2.2: Backend — Log Corrections to `audit_wal`

**Status:** ⬜ Not Started

**Current State:**
- [audit_wal.md](file:///home/colton/Desktop/FutureBuild_HQ/XUI/backend/shadow/internal/chat/audit_wal.md) — placeholder stub, no Go implementation

**Required Implementation:**

1. **Create `backend/internal/api/handlers/correction_handler.go`** [NEW]:
   ```go
   // POST /api/v1/corrections
   func HandleCorrectionBatch(w http.ResponseWriter, r *http.Request) {
       var events []CorrectionEvent
       // Decode, validate, log
   }
   ```

2. **Create `backend/internal/audit/wal.go`** [NEW]:
   ```go
   // Write-Ahead Log for correction events
   type AuditWAL struct {
       db *sql.DB // or file-based JSON-lines
   }
   
   type CorrectionEntry struct {
       ID                string    `json:"id"`
       ArtifactID        string    `json:"artifact_id"`
       ArtifactType      string    `json:"artifact_type"`
       FieldPath         string    `json:"field_path"`
       OldValue          any       `json:"old_value"`
       NewValue          any       `json:"new_value"`
       OriginalConfidence float64  `json:"original_confidence"`
       UserID            string    `json:"user_id"`
       Timestamp         time.Time `json:"timestamp"`
   }
   
   func (w *AuditWAL) Append(entry CorrectionEntry) error
   ```

3. **Storage options** (in priority order):
   - **PostgreSQL table** `correction_log` — structured, queryable
   - **JSON-lines file** — simple append-only, good for later batch processing
   - **Both** — write to DB + emit to event stream for async processing

4. **Future use:** This data feeds back into VisionAgent prompt refinement:
   - "Users frequently correct 'unit_price_cents' on electrical invoices — improve extraction accuracy for this field."

---

## Codebase References

| File | Path | Status | Notes |
|------|------|--------|-------|
| fb-artifact-invoice.ts | `frontend/src/components/artifacts/fb-artifact-invoice.ts` | Modify | Add correction capture in `_save()` |
| audit_wal.md | `backend/shadow/internal/chat/audit_wal.md` | Stub | Needs Go implementation |
| correction_handler.go | `backend/internal/api/handlers/` | [NEW] | API endpoint for corrections |
| wal.go | `backend/internal/audit/` | [NEW] | Write-ahead log for corrections |

## Verification Plan

- **Manual:** Edit an AI-extracted invoice field → save → verify correction event logged in backend (check DB or logs)
- **Manual:** Edit a manually-entered field → save → verify NO correction event (not AI-sourced)
- **Automated:** API test: `POST /api/v1/corrections` with valid payload → verify 200 response and DB entry
- **Automated:** API test: `POST /api/v1/corrections` with invalid payload → verify 400 response
