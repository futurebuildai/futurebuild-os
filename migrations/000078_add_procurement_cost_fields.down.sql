ALTER TABLE procurement_items
    DROP COLUMN IF EXISTS estimated_cost_cents,
    DROP COLUMN IF EXISTS vendor,
    DROP COLUMN IF EXISTS brand,
    DROP COLUMN IF EXISTS model,
    DROP COLUMN IF EXISTS sku;
