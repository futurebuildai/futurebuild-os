-- L7 Fix: Rollback schema drift fixes
DROP TRIGGER IF EXISTS update_projects_modtime ON projects;
ALTER TABLE projects DROP COLUMN IF EXISTS updated_at;
ALTER TABLE procurement_items DROP COLUMN IF EXISTS expected_delivery_date;
ALTER TABLE project_tasks DROP COLUMN IF EXISTS override_reason;

