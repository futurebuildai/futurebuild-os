-- Project Completion Lifecycle
-- Adds completion tracking to projects and a dedicated completion_reports table.

-- Add completion columns to projects
ALTER TABLE projects
    ADD COLUMN completed_at TIMESTAMPTZ,
    ADD COLUMN completed_by UUID REFERENCES users(id);

-- Completion reports: one per project, generated at completion time
CREATE TABLE completion_reports (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    generated_by    UUID REFERENCES users(id),
    schedule_summary    JSONB NOT NULL,
    budget_summary      JSONB NOT NULL,
    weather_impact_summary JSONB,
    procurement_summary    JSONB,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One report per project
CREATE UNIQUE INDEX idx_completion_reports_project_id ON completion_reports(project_id);
