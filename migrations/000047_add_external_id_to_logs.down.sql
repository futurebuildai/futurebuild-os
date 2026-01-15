-- Rollback: Remove external_id column
-- See PRODUCTION_PLAN.md Step 48

DROP INDEX IF EXISTS idx_communication_logs_external_id;
ALTER TABLE communication_logs DROP COLUMN IF EXISTS external_id;
