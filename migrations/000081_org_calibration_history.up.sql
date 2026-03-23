-- Migration: Org-Level Calibration History
-- Stores actual vs predicted duration ratios per WBS code per org.
-- Used to train org-specific duration multipliers that improve schedule accuracy.

CREATE TABLE IF NOT EXISTS calibration_entries (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID NOT NULL,
    project_id    UUID NOT NULL,
    wbs_code      TEXT NOT NULL,
    predicted_days DOUBLE PRECISION NOT NULL,
    actual_days    DOUBLE PRECISION NOT NULL,
    ratio          DOUBLE PRECISION NOT NULL, -- actual / predicted
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Prevent duplicate calibration for same project+task
    UNIQUE (project_id, wbs_code)
);

-- Primary lookup: org multipliers by WBS code
CREATE INDEX idx_calibration_org_wbs ON calibration_entries (org_id, wbs_code);

-- Cleanup/audit: find entries by project
CREATE INDEX idx_calibration_project ON calibration_entries (project_id);
