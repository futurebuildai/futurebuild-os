-- Migration: Restore is_inspection column to project_tasks
-- Required for Event Duration Locking per BACKEND_SCOPE.md Section 4.2
-- See CPM_RES_MODEL_SPEC.md Section 11.2.1

ALTER TABLE project_tasks ADD COLUMN IF NOT EXISTS is_inspection BOOLEAN DEFAULT FALSE;

-- Add comment for documentation
COMMENT ON COLUMN project_tasks.is_inspection IS 'Critical for Event Duration Locking rule - inspections have fixed SAF=1.0';

-- Create index for query performance when filtering inspections
CREATE INDEX IF NOT EXISTS idx_project_tasks_is_inspection ON project_tasks(is_inspection) WHERE is_inspection = TRUE;
