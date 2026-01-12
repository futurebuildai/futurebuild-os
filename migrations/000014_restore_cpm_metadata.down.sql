-- Migration: Down - Restore CPM Metadata
DROP TRIGGER IF EXISTS update_project_tasks_modtime ON project_tasks;
ALTER TABLE project_tasks 
    DROP COLUMN IF EXISTS early_start,
    DROP COLUMN IF EXISTS early_finish,
    DROP COLUMN IF EXISTS late_start,
    DROP COLUMN IF EXISTS late_finish,
    DROP COLUMN IF EXISTS total_float_days,
    DROP COLUMN IF EXISTS is_on_critical_path,
    DROP COLUMN IF EXISTS updated_at;

ALTER TABLE project_tasks RENAME COLUMN calculated_duration TO calculated_duration_days;
ALTER TABLE project_tasks RENAME COLUMN weather_adjusted_duration TO weather_adjusted_duration_days;
