-- Rollback: Convert cents back to DECIMAL
-- WARNING: Precision loss possible if fractional cents exist

BEGIN;

-- =============================================================================
-- INVOICES TABLE
-- =============================================================================

ALTER TABLE invoices ADD COLUMN amount DECIMAL(15,2);
UPDATE invoices SET amount = amount_cents / 100.0;
ALTER TABLE invoices DROP COLUMN amount_cents;
ALTER TABLE invoices ALTER COLUMN amount SET NOT NULL;
ALTER TABLE invoices ALTER COLUMN amount SET DEFAULT 0.0;

-- =============================================================================
-- PROJECT BUDGETS TABLE
-- =============================================================================

ALTER TABLE project_budgets 
    ADD COLUMN estimated_amount DECIMAL(15,2) DEFAULT 0.0,
    ADD COLUMN committed_amount DECIMAL(15,2) DEFAULT 0.0,
    ADD COLUMN actual_amount DECIMAL(15,2) DEFAULT 0.0;

UPDATE project_budgets SET 
    estimated_amount = estimated_amount_cents / 100.0,
    committed_amount = committed_amount_cents / 100.0,
    actual_amount = actual_amount_cents / 100.0;

ALTER TABLE project_budgets 
    DROP COLUMN estimated_amount_cents,
    DROP COLUMN committed_amount_cents,
    DROP COLUMN actual_amount_cents;

ALTER TABLE project_budgets ALTER COLUMN estimated_amount SET NOT NULL;
ALTER TABLE project_budgets ALTER COLUMN committed_amount SET NOT NULL;
ALTER TABLE project_budgets ALTER COLUMN actual_amount SET NOT NULL;

COMMIT;
