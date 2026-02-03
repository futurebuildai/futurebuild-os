-- Rollback: Project Completion Lifecycle

DROP TABLE IF EXISTS completion_reports;

ALTER TABLE projects
    DROP COLUMN IF EXISTS completed_at,
    DROP COLUMN IF EXISTS completed_by;
