-- Revert: remove enriched invoice columns
ALTER TABLE invoices DROP COLUMN IF EXISTS confidence;
ALTER TABLE invoices DROP COLUMN IF EXISTS invoice_date;
ALTER TABLE invoices DROP COLUMN IF EXISTS invoice_number;
