-- Step 84: Field Feedback Loop
-- Create project_assets table for tracking uploaded photos and their vision analysis status.
-- See STEP_84_FIELD_FEEDBACK.md Section 2

CREATE TYPE analysis_status_type AS ENUM ('processing', 'completed', 'failed');

CREATE TABLE project_assets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    task_id         UUID REFERENCES project_tasks(id) ON DELETE SET NULL,
    uploaded_by     TEXT NOT NULL,
    file_name       TEXT NOT NULL,
    file_url        TEXT NOT NULL,
    mime_type       TEXT NOT NULL,
    file_size_bytes BIGINT NOT NULL DEFAULT 0,
    analysis_status analysis_status_type NOT NULL DEFAULT 'processing',
    analysis_result JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_project_assets_project_id ON project_assets(project_id);
CREATE INDEX idx_project_assets_task_id ON project_assets(task_id);
CREATE INDEX idx_project_assets_analysis_status ON project_assets(analysis_status);
