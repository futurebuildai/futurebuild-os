-- Revert: drop documents table
DROP INDEX IF EXISTS idx_documents_project_id;
DROP TABLE IF EXISTS documents;
