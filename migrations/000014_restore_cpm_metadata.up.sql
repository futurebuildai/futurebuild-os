-- Migration: Restore CPM Metadata
-- Correcting aggressive purge from 000006

ALTER TABLE project_tasks 
    ADD COLUMN IF NOT EXISTS early_start TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS early_finish TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS late_start TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS late_finish TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS total_float_days FLOAT DEFAULT 0.0,
    ADD COLUMN IF NOT EXISTS is_on_critical_path BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'project_tasks' AND column_name = 'calculated_duration_days') THEN
        ALTER TABLE project_tasks RENAME COLUMN calculated_duration_days TO calculated_duration;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'project_tasks' AND column_name = 'weather_adjusted_duration_days') THEN
        ALTER TABLE project_tasks RENAME COLUMN weather_adjusted_duration_days TO weather_adjusted_duration;
    END IF;
END $$;

-- Restore trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_project_tasks_modtime ON project_tasks;
CREATE TRIGGER update_project_tasks_modtime BEFORE UPDATE ON project_tasks FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
