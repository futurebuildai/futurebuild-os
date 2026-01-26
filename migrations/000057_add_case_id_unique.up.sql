-- Idempotency constraint for Automated PR Review
-- See docs/AUTOMATED_PR_REVIEW_PRD.md
--
-- Creates a unique index on case_id to prevent duplicate reviews for the same
-- PR SHA. GitHub PR format: GH_{owner}/{repo}#{number}_{sha}

CREATE UNIQUE INDEX IF NOT EXISTS idx_tribunal_decisions_case_id_unique
ON tribunal_decisions (case_id)
WHERE case_id IS NOT NULL;

COMMENT ON COLUMN tribunal_decisions.case_id IS
'Unique case identifier. GitHub PR format: GH_{owner}/{repo}#{number}_{sha}';
