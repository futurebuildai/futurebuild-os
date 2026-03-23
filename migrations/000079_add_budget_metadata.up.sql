-- Budget seeding metadata and regional cost multipliers
-- Supports AI-generated budget estimates with confidence scoring

BEGIN;

-- Budget source tracking and locking
ALTER TABLE project_budgets
    ADD COLUMN IF NOT EXISTS source material_source_type DEFAULT 'default',
    ADD COLUMN IF NOT EXISTS confidence FLOAT DEFAULT 0.0,
    ADD COLUMN IF NOT EXISTS is_locked BOOLEAN DEFAULT FALSE;

-- Regional cost multiplier stored per-project in existing context table
ALTER TABLE project_context
    ADD COLUMN IF NOT EXISTS regional_cost_multiplier FLOAT DEFAULT 1.0,
    ADD COLUMN IF NOT EXISTS cost_index_region VARCHAR(100);

-- Prevent duplicate budget rows per project+phase
CREATE UNIQUE INDEX IF NOT EXISTS idx_project_budgets_project_phase
    ON project_budgets(project_id, wbs_phase_id);

COMMIT;
