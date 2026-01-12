-- Rollback migration for is_human_review_required
ALTER TABLE invoices DROP COLUMN is_human_review_required;
ALTER TABLE project_tasks DROP COLUMN is_human_review_required;
