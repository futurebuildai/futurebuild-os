-- Migration: Add Invoices and Refine Financials
-- Domain 3: Financials (The Wallet) per DATA_SPINE_SPEC.md

-- Create Invoice Status Enum
CREATE TYPE invoice_status_type AS ENUM ('Pending', 'Approved', 'Exported');

-- 4.2 INVOICES
CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE NOT NULL,
    vendor_name VARCHAR(255) NOT NULL,
    amount DECIMAL(15,2) DEFAULT 0.0 NOT NULL,
    line_items JSONB DEFAULT '[]' NOT NULL,
    detected_wbs_code VARCHAR(50),
    status invoice_status_type DEFAULT 'Pending' NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for Invoices
CREATE INDEX IF NOT EXISTS idx_invoices_project_id ON invoices(project_id);
CREATE INDEX IF NOT EXISTS idx_invoices_vendor_name ON invoices(vendor_name);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);

-- Ensure project_budgets table matches SPEC exactly if not already done in init
-- (It was stubbed in 000001, but we ensure updated_at or other missing fields)
ALTER TABLE project_budgets ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE project_budgets ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- Trigger to update updated_at if it doesn't exist
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_invoices_modtime BEFORE UPDATE ON invoices FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_project_budgets_modtime BEFORE UPDATE ON project_budgets FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
