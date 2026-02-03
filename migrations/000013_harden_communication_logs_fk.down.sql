-- Revert: restore CASCADE on communication_logs FK
ALTER TABLE communication_logs
DROP CONSTRAINT IF EXISTS communication_logs_project_id_fkey;

ALTER TABLE communication_logs
ADD CONSTRAINT communication_logs_project_id_fkey
FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;
