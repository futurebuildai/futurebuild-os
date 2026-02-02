-- Step 82: Add Draft and Rejected invoice statuses for the Action Loop.
-- See PHASE_13_PRD.md Section 3.1 and STEP_82_INTERACTIVE_INVOICE.md
--
-- Draft: AI-generated invoices start here, editable by users.
-- Rejected: User explicitly rejects an invoice (Step 83).
--
-- PostgreSQL ENUM types require ALTER TYPE ... ADD VALUE.
-- These are idempotent (IF NOT EXISTS prevents errors on re-run).

ALTER TYPE invoice_status_type ADD VALUE IF NOT EXISTS 'Draft' BEFORE 'Pending';
ALTER TYPE invoice_status_type ADD VALUE IF NOT EXISTS 'Rejected' AFTER 'Approved';

-- Migrate existing 'Pending' invoices to 'Draft' so they become editable.
-- This is the correct semantic: AI-generated invoices should start as Draft,
-- not Pending (which implies human review has occurred).
UPDATE invoices SET status = 'Draft' WHERE status = 'Pending';
