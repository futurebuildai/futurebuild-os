-- Migration: Add Outbox Status to Communication Logs
-- Purpose: Enable Transactional Outbox pattern for at-most-once delivery
-- Reference: P0 Fix for SubLiaisonAgent non-atomic side effects

-- 1. Create status enum for outbox pattern
CREATE TYPE communication_send_status AS ENUM ('PENDING', 'SENT', 'FAILED');

-- 2. Add status column (default SENT for backwards compat with existing records)
ALTER TABLE communication_logs 
ADD COLUMN send_status communication_send_status DEFAULT 'SENT';

-- 3. Partial index for outbox polling (find stuck PENDING records efficiently)
CREATE INDEX idx_communication_logs_outbox 
ON communication_logs(send_status, timestamp) 
WHERE send_status = 'PENDING';
