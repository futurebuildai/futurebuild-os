-- Migration: Add external_id for webhook idempotency
-- See PRODUCTION_PLAN.md Step 48 (Idempotency Amendment)
-- Webhook providers (Twilio/SendGrid) guarantee at-least-once delivery.
-- This unique index enforces idempotency at the database level.

ALTER TABLE communication_logs 
ADD COLUMN external_id VARCHAR(255);

-- Unique index for idempotency (only on non-null values)
-- Allows "ON CONFLICT DO NOTHING" or cheap unique violation handling
CREATE UNIQUE INDEX idx_communication_logs_external_id 
ON communication_logs(external_id) 
WHERE external_id IS NOT NULL;
