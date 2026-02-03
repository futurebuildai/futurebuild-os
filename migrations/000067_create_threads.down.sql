-- 000067_create_threads.down.sql
-- Reverse thread support migration.

BEGIN;

-- Drop FK and index from chat_messages
ALTER TABLE chat_messages DROP CONSTRAINT IF EXISTS fk_chat_messages_thread_id;
DROP INDEX IF EXISTS idx_chat_messages_thread_id_created_at;
ALTER TABLE chat_messages DROP COLUMN IF EXISTS thread_id;

-- Drop threads table and indexes
DROP INDEX IF EXISTS idx_threads_project_general;
DROP INDEX IF EXISTS idx_threads_project_id;
DROP TABLE IF EXISTS threads;

COMMIT;
