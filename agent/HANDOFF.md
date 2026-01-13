# Handoff: Phase 5 Step 42

**Previous Step:** 41 (Document Re-processing & Audit Trail) - **COMPLETED**
**Current Step:** 42 (Mock Ingestion Pipeline)

## Status
- **Step 41 Complete**: Implemented document re-processing with audit trail.
  - Added `source_document_id` to `invoices` table.
  - Added `updated_at`, `reprocessed_count` to `documents` table.
  - Implemented `ReprocessDocument` handler with chained Re-extraction -> UPSERT.
  - Verified with new integration test `test/integration/reprocess_test.go`.
- **Ready for Step 42**: Need to create a test fixture for "perfect" JSON injection to verify DB logic in isolation from AI.

## Context for Step 42
The goal is to create a deterministic test fixture that simulates the AI's output. This allows us to test the `InvoiceService.SaveExtraction` logic (mapping, UPSERT, Review Flags) without relying on the actual Vertex AI calls or mocks that might drift. This "Mock Ingestion Pipeline" will serve as a regression suite for the database layer.

## Key Files
- `test/fixtures/perfect_invoice.json` (To be created)
- `test/integration/pipeline_test.go` (To be created)
- `internal/service/invoice_service.go` (Target of testing)

## Next Actions
1.  Create `test/fixtures` directory.
2.  Add `perfect_invoice.json` matching `types.InvoiceExtraction`.
3.  Create `test/integration/pipeline_test.go` that loads this JSON and calls `invoiceService.SaveExtraction`.
4.  Assert database state matches the JSON exactly.
