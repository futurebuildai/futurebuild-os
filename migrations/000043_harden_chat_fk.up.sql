-- CTO-001 Remediation: Harden chat_messages FK policy
-- Change ON DELETE CASCADE to ON DELETE RESTRICT to preserve audit trail.
-- See CTO Audit Report (Step 43.6 Review)

-- Drop existing constraints
ALTER TABLE chat_messages
  DROP CONSTRAINT IF EXISTS chat_messages_project_id_fkey;

ALTER TABLE chat_messages
  DROP CONSTRAINT IF EXISTS chat_messages_user_id_fkey;

-- Re-add with RESTRICT policy
ALTER TABLE chat_messages
  ADD CONSTRAINT chat_messages_project_id_fkey
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE RESTRICT;

ALTER TABLE chat_messages
  ADD CONSTRAINT chat_messages_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
