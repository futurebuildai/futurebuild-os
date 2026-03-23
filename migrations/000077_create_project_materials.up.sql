-- Material list extraction and quantity takeoffs
-- See PROJECT_ONBOARDING_REPORT.md: Tier 1 Procurement Timeline Integration
-- All monetary values stored as BIGINT cents (IEEE 754 prevention)

BEGIN;

CREATE TYPE material_source_type AS ENUM ('ai', 'user', 'default');

CREATE TABLE IF NOT EXISTS project_materials (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    wbs_phase_code VARCHAR(20) NOT NULL,
    name VARCHAR(500) NOT NULL,
    category VARCHAR(100) NOT NULL,
    quantity FLOAT DEFAULT 0.0,
    unit VARCHAR(50) NOT NULL DEFAULT 'ea',
    unit_cost_cents BIGINT NOT NULL DEFAULT 0,
    total_cost_cents BIGINT NOT NULL DEFAULT 0,
    source material_source_type NOT NULL DEFAULT 'default',
    confidence FLOAT DEFAULT 0.0,
    brand VARCHAR(255),
    model VARCHAR(255),
    sku VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Upsert support: prevent duplicate material per project+phase+name
CREATE UNIQUE INDEX IF NOT EXISTS idx_project_materials_upsert
    ON project_materials(project_id, wbs_phase_code, name);

CREATE INDEX idx_project_materials_project_id ON project_materials(project_id);
CREATE INDEX idx_project_materials_wbs_phase ON project_materials(wbs_phase_code);
CREATE INDEX idx_project_materials_category ON project_materials(category);

CREATE TRIGGER update_project_materials_modtime
    BEFORE UPDATE ON project_materials
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

COMMIT;
