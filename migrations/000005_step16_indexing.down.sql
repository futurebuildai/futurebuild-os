-- Migration Down: Step 16 Indexing
DROP INDEX IF EXISTS idx_projects_status;
DROP INDEX IF EXISTS idx_projects_name;
