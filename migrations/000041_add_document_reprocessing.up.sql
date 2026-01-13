-- Migration: Add Document Re-processing Audit Trail
-- See PRODUCTION_PLAN.md Step 41
-- See DATA_SPINE_SPEC.md Section 4.2

-- Add audit trail columns to documents table
ALTER TABLE documents 
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS reprocessed_count INT DEFAULT 0;

-- Create auto-update trigger for documents (using existing function from migration 000002)
DROP TRIGGER IF EXISTS update_documents_modtime ON documents;
CREATE TRIGGER update_documents_modtime 
    BEFORE UPDATE ON documents 
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- Link invoices to source documents for re-processing version tracking
-- Uses SET NULL to preserve invoice history if document is deleted
ALTER TABLE invoices 
    ADD COLUMN IF NOT EXISTS source_document_id UUID REFERENCES documents(id) ON DELETE SET NULL;

-- Index for invoice-document lookups (enables efficient UPSERT checks)
CREATE INDEX IF NOT EXISTS idx_invoices_source_document_id ON invoices(source_document_id);
