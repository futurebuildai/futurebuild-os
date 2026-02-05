-- Rollback portal schema enhancements.

ALTER TABLE threads DROP COLUMN IF EXISTS thread_type;
DROP TYPE IF EXISTS thread_type_enum;

DROP INDEX IF EXISTS idx_chat_messages_contact_id;
ALTER TABLE chat_messages DROP CONSTRAINT IF EXISTS chk_chat_sender;
ALTER TABLE chat_messages ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE chat_messages DROP COLUMN IF EXISTS contact_id;
