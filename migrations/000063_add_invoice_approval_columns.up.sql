-- Migrate existing 'Pending' invoices to 'Draft' (moved from migration 62
-- because PostgreSQL cannot use new enum values in the same transaction).
UPDATE invoices SET status = 'Draft' WHERE status = 'Pending';

-- Step 83: Add approval/rejection metadata columns to invoices.
-- See STEP_83_APPROVAL_ACTIONS.md Section 2.3
--
-- Records who approved/rejected and when, plus rejection reason.
-- These create an audit trail directly on the invoice record.

ALTER TABLE invoices ADD COLUMN IF NOT EXISTS approved_by_id VARCHAR(255);
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS rejected_by_id VARCHAR(255);
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMPTZ;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS rejection_reason TEXT;
