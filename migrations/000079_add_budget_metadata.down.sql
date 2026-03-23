BEGIN;

DROP INDEX IF EXISTS idx_project_budgets_project_phase;

ALTER TABLE project_context
    DROP COLUMN IF EXISTS regional_cost_multiplier,
    DROP COLUMN IF EXISTS cost_index_region;

ALTER TABLE project_budgets
    DROP COLUMN IF EXISTS source,
    DROP COLUMN IF EXISTS confidence,
    DROP COLUMN IF EXISTS is_locked;

COMMIT;
