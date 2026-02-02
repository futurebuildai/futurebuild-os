-- Down migration: Remove approval columns.
ALTER TABLE invoices DROP COLUMN IF EXISTS approved_by_id;
ALTER TABLE invoices DROP COLUMN IF EXISTS approved_at;
ALTER TABLE invoices DROP COLUMN IF EXISTS rejected_by_id;
ALTER TABLE invoices DROP COLUMN IF EXISTS rejected_at;
ALTER TABLE invoices DROP COLUMN IF EXISTS rejection_reason;
