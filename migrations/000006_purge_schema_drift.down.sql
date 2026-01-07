-- Rollback: Purge Schema Drift
-- Re-adds columns for rollback purposes only

-- 1. Re-add zombie columns to project_tasks
ALTER TABLE project_tasks ADD COLUMN IF NOT EXISTS is_inspection BOOLEAN DEFAULT FALSE;
ALTER TABLE project_tasks ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE project_tasks ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- 2. Re-add user_id to communication_logs
ALTER TABLE communication_logs ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE SET NULL;

-- 3. Re-add wbs_phase_id as UUID (reverting to 000004 state)
ALTER TABLE project_assignments DROP COLUMN IF EXISTS wbs_phase_id;
ALTER TABLE project_assignments ADD COLUMN wbs_phase_id UUID REFERENCES wbs_phases(id) ON DELETE CASCADE;
