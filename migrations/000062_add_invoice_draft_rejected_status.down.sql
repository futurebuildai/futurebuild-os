-- Down migration: Revert Draft invoices back to Pending.
-- Note: PostgreSQL does not support removing values from ENUM types.
-- The enum values Draft and Rejected will remain but become unused.
UPDATE invoices SET status = 'Pending' WHERE status = 'Draft';
UPDATE invoices SET status = 'Pending' WHERE status = 'Rejected';
