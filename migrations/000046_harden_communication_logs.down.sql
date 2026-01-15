-- Rollback: Harden Communication Logs

DROP INDEX IF EXISTS idx_communication_logs_related_entity;
DROP INDEX IF EXISTS idx_communication_logs_timestamp;

ALTER TABLE communication_logs 
DROP COLUMN related_entity_type,
DROP COLUMN related_entity_id;
