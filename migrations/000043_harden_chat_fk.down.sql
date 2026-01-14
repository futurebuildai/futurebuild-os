-- Rollback: Restore original CASCADE policy

ALTER TABLE chat_messages
  DROP CONSTRAINT IF EXISTS chat_messages_project_id_fkey;

ALTER TABLE chat_messages
  DROP CONSTRAINT IF EXISTS chat_messages_user_id_fkey;

ALTER TABLE chat_messages
  ADD CONSTRAINT chat_messages_project_id_fkey
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;

ALTER TABLE chat_messages
  ADD CONSTRAINT chat_messages_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
