-- Migration: Purge Schema Drift
-- Aligns schema with DATA_SPINE_SPEC.md and Go models

-- 1. Drop zombie columns from project_tasks
-- These were added in 000004 but are NOT in DATA_SPINE_SPEC.md Section 3.3
ALTER TABLE project_tasks DROP COLUMN IF EXISTS is_inspection;
ALTER TABLE project_tasks DROP COLUMN IF EXISTS created_at;
ALTER TABLE project_tasks DROP COLUMN IF EXISTS updated_at;

-- Drop the trigger that was created for updated_at
DROP TRIGGER IF EXISTS update_project_tasks_modtime ON project_tasks;

-- 2. Drop user_id from communication_logs
-- DATA_SPINE_SPEC.md Section 5.1 does NOT include user_id
ALTER TABLE communication_logs DROP COLUMN IF EXISTS user_id;

-- 3. Fix wbs_phase_id in project_assignments
-- DATA_SPINE_SPEC.md Section 3.5 defines this as VARCHAR, not UUID
-- First drop the FK constraint if it exists
ALTER TABLE project_assignments DROP CONSTRAINT IF EXISTS project_assignments_wbs_phase_id_fkey;

-- Then drop the old column and recreate as VARCHAR
ALTER TABLE project_assignments DROP COLUMN IF EXISTS wbs_phase_id;
ALTER TABLE project_assignments ADD COLUMN wbs_phase_id VARCHAR(20) NOT NULL DEFAULT '0.0';
