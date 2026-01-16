-- Migration: Rollback Outbox Status
-- Removes send_status column and enum from communication_logs

DROP INDEX IF EXISTS idx_communication_logs_outbox;
ALTER TABLE communication_logs DROP COLUMN IF EXISTS send_status;
DROP TYPE IF EXISTS communication_send_status;
