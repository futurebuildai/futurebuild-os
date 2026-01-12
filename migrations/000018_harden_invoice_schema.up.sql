-- Add enriched fields to invoices table for Step 37 Audit Remediation
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS confidence FLOAT;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS invoice_date DATE;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS invoice_number VARCHAR(100);
