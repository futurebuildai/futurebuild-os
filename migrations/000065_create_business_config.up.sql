-- Step 87: Business Config for Construction Physics tuning.
-- See STEP_87_CONFIG_PERSISTENCE.md Section 1
-- One config per organization (tenant). Stores speed multiplier and work days.
CREATE TABLE business_config (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id           UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    speed_multiplier DECIMAL(3, 2) NOT NULL DEFAULT 1.00,
    work_days        JSONB NOT NULL DEFAULT '[1, 2, 3, 4, 5]'::jsonb,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_business_config_org_id UNIQUE (org_id),
    -- C-2: Range aligned with frontend slider (0.5-1.5)
    CONSTRAINT chk_speed_multiplier_range CHECK (speed_multiplier >= 0.50 AND speed_multiplier <= 1.50)
);
-- H-4: Redundant explicit index removed — UNIQUE constraint on org_id already creates an index.
