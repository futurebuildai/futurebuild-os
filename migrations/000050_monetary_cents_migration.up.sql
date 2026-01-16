-- Migration: Convert monetary values from DECIMAL to BIGINT (cents)
-- This migration safely converts all monetary values to integer cents representation
-- to prevent IEEE 754 floating-point precision errors.

BEGIN;

-- =============================================================================
-- INVOICES TABLE
-- =============================================================================

-- 1. Add new cents column
ALTER TABLE invoices ADD COLUMN amount_cents BIGINT;

-- 2. Convert existing data (DECIMAL * 100 → BIGINT)
UPDATE invoices SET amount_cents = ROUND(amount * 100)::BIGINT WHERE amount IS NOT NULL;

-- 3. Drop old column and apply constraints
ALTER TABLE invoices DROP COLUMN amount;
ALTER TABLE invoices ALTER COLUMN amount_cents SET NOT NULL;
ALTER TABLE invoices ALTER COLUMN amount_cents SET DEFAULT 0;

-- =============================================================================
-- PROJECT BUDGETS TABLE
-- =============================================================================

-- 1. Add new cents columns
ALTER TABLE project_budgets 
    ADD COLUMN estimated_amount_cents BIGINT DEFAULT 0,
    ADD COLUMN committed_amount_cents BIGINT DEFAULT 0,
    ADD COLUMN actual_amount_cents BIGINT DEFAULT 0;

-- 2. Convert existing data
UPDATE project_budgets SET 
    estimated_amount_cents = ROUND(estimated_amount * 100)::BIGINT,
    committed_amount_cents = ROUND(committed_amount * 100)::BIGINT,
    actual_amount_cents = ROUND(actual_amount * 100)::BIGINT
WHERE estimated_amount IS NOT NULL 
   OR committed_amount IS NOT NULL 
   OR actual_amount IS NOT NULL;

-- 3. Drop old columns
ALTER TABLE project_budgets 
    DROP COLUMN estimated_amount,
    DROP COLUMN committed_amount,
    DROP COLUMN actual_amount;

-- 4. Apply NOT NULL constraints after data migration
ALTER TABLE project_budgets ALTER COLUMN estimated_amount_cents SET NOT NULL;
ALTER TABLE project_budgets ALTER COLUMN committed_amount_cents SET NOT NULL;
ALTER TABLE project_budgets ALTER COLUMN actual_amount_cents SET NOT NULL;

COMMIT;
