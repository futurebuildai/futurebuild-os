-- Add planned/actual date columns to project_tasks for in-progress project support.
-- The Go model (internal/models/project_task.go) defines these fields but they were
-- never added to the DB schema (schema drift from enterprise remediation).
ALTER TABLE project_tasks ADD COLUMN IF NOT EXISTS planned_start TIMESTAMPTZ;
ALTER TABLE project_tasks ADD COLUMN IF NOT EXISTS planned_end TIMESTAMPTZ;
ALTER TABLE project_tasks ADD COLUMN IF NOT EXISTS actual_start TIMESTAMPTZ;
ALTER TABLE project_tasks ADD COLUMN IF NOT EXISTS actual_end TIMESTAMPTZ;
