-- Rollback: Remove Invoices and Financial Refinements

DROP TRIGGER IF EXISTS update_project_budgets_modtime ON project_budgets;
DROP TRIGGER IF EXISTS update_invoices_modtime ON invoices;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS invoices;
DROP TYPE IF EXISTS invoice_status_type;

-- Removing columns from project_budgets added in migration 2
ALTER TABLE project_budgets DROP COLUMN IF EXISTS updated_at;
ALTER TABLE project_budgets DROP COLUMN IF EXISTS created_at;
