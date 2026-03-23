-- Add cost and specification fields to procurement_items
-- Enables cost tracking on long-lead items for budget integration

ALTER TABLE procurement_items
    ADD COLUMN IF NOT EXISTS estimated_cost_cents BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS vendor VARCHAR(255),
    ADD COLUMN IF NOT EXISTS brand VARCHAR(255),
    ADD COLUMN IF NOT EXISTS model VARCHAR(255),
    ADD COLUMN IF NOT EXISTS sku VARCHAR(255);
