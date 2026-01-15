-- Migration: Harden Communication Logs
-- Purpose: Resolve Code Review Warns #1 (Indexing) and #2 (Fragile Querying)
-- Reference: Step 46 /codereview feedback

-- 1. Add specific targeting columns
ALTER TABLE communication_logs 
ADD COLUMN related_entity_id UUID,
ADD COLUMN related_entity_type VARCHAR(50);

-- 2. Add Index for Dampening Checks (Warn #1)
-- Timestamp filters are frequent for "last X hours" checks
CREATE INDEX idx_communication_logs_timestamp ON communication_logs(timestamp);

-- 3. Add Index for Entity Lookups (Warn #2)
CREATE INDEX idx_communication_logs_related_entity ON communication_logs(related_entity_id);
