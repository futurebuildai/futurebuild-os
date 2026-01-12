-- Migration: Harden Communication Logs FK
-- Prevents audit trail vaporization per CTO Audit Step 31

ALTER TABLE communication_logs 
DROP CONSTRAINT IF EXISTS communication_logs_project_id_fkey;

ALTER TABLE communication_logs 
ADD CONSTRAINT communication_logs_project_id_fkey 
FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE RESTRICT;
