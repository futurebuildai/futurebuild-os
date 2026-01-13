-- Rollback Migration: Remove Document Re-processing Audit Trail
-- See PRODUCTION_PLAN.md Step 41

DROP TRIGGER IF EXISTS update_documents_modtime ON documents;
DROP INDEX IF EXISTS idx_invoices_source_document_id;
ALTER TABLE invoices DROP COLUMN IF EXISTS source_document_id;
ALTER TABLE documents DROP COLUMN IF EXISTS updated_at;
ALTER TABLE documents DROP COLUMN IF EXISTS reprocessed_count;
