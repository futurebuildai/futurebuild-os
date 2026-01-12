-- Rollback: Remove is_inspection column from project_tasks

DROP INDEX IF EXISTS idx_project_tasks_is_inspection;
ALTER TABLE project_tasks DROP COLUMN IF EXISTS is_inspection;
