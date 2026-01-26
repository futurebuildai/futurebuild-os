-- Rollback: Remove unique constraint on case_id
-- See docs/AUTOMATED_PR_REVIEW_PRD.md

DROP INDEX IF EXISTS idx_tribunal_decisions_case_id_unique;
