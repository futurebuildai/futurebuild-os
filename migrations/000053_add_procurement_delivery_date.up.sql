-- L7 Fix: Add missing columns referenced by schedule_service.go
-- This migration fixes schema drift discovered during integration testing
-- See PRODUCTION_PLAN.md Step 62.2.5 (L7 Remediation & Hardening)

-- 1. Add expected_delivery_date to procurement_items (for getMaterialConstraints)
ALTER TABLE procurement_items 
ADD COLUMN IF NOT EXISTS expected_delivery_date TIMESTAMPTZ;

-- 2. Add updated_at to projects table (for RecalculateSchedule)
ALTER TABLE projects 
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- 3. Create trigger to auto-update the timestamp
CREATE TRIGGER update_projects_modtime 
BEFORE UPDATE ON projects 
FOR EACH ROW 
EXECUTE PROCEDURE update_updated_at_column();

-- 4. Add override_reason to project_tasks (for UpdateTaskDuration)
ALTER TABLE project_tasks 
ADD COLUMN IF NOT EXISTS override_reason TEXT;
