-- Migration: Fix Project Task Drift
-- Aligning with BACKEND_SCOPE.md Authority

-- 1. Update project_tasks table
ALTER TABLE project_tasks ADD COLUMN IF NOT EXISTS is_inspection BOOLEAN DEFAULT FALSE;
ALTER TABLE project_tasks ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE project_tasks ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- 2. Update project_assignments table
-- First, drop the old column (we assume local dev, so data loss is acceptable for this drift correction)
-- If this was production, we would need a mapping table.
ALTER TABLE project_assignments DROP COLUMN IF EXISTS wbs_phase_id;
ALTER TABLE project_assignments ADD COLUMN wbs_phase_id UUID REFERENCES wbs_phases(id) ON DELETE CASCADE;

-- 3. Update project_assignments contact_id deletion policy
ALTER TABLE project_assignments DROP CONSTRAINT IF EXISTS project_assignments_contact_id_fkey;
ALTER TABLE project_assignments ADD CONSTRAINT project_assignments_contact_id_fkey 
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE SET NULL;

-- 4. Add Update Trigger for project_tasks
CREATE TRIGGER update_project_tasks_modtime BEFORE UPDATE ON project_tasks FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
