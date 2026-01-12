-- Add review flags to invoices and project_tasks
-- See DATA_SPINE_SPEC.md Section 4.2
ALTER TABLE invoices ADD COLUMN is_human_review_required BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE project_tasks ADD COLUMN is_human_review_required BOOLEAN NOT NULL DEFAULT FALSE;
