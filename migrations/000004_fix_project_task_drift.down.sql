-- Migration: Rollback Fix Project Task Drift

ALTER TABLE project_tasks DROP COLUMN IF EXISTS is_inspection;
ALTER TABLE project_tasks DROP COLUMN IF EXISTS created_at;
ALTER TABLE project_tasks DROP COLUMN IF EXISTS updated_at;

DROP TRIGGER IF EXISTS update_project_tasks_modtime ON project_tasks;

ALTER TABLE project_assignments DROP COLUMN IF EXISTS wbs_phase_id;
ALTER TABLE project_assignments ADD COLUMN wbs_phase_id VARCHAR(20);

ALTER TABLE project_assignments DROP CONSTRAINT IF EXISTS project_assignments_contact_id_fkey;
ALTER TABLE project_assignments ADD CONSTRAINT project_assignments_contact_id_fkey 
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE;
