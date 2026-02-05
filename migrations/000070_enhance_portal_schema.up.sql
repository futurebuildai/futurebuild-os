-- Portal schema enhancements: allow contacts as chat senders, distinguish portal threads.
-- See Client & Subcontractor Portal backend plan.

-- chat_messages: allow portal contacts as senders
ALTER TABLE chat_messages ADD COLUMN contact_id UUID REFERENCES contacts(id) ON DELETE SET NULL;
ALTER TABLE chat_messages ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE chat_messages ADD CONSTRAINT chk_chat_sender
    CHECK (user_id IS NOT NULL OR contact_id IS NOT NULL);
CREATE INDEX idx_chat_messages_contact_id ON chat_messages(contact_id) WHERE contact_id IS NOT NULL;

-- threads: distinguish portal threads from internal threads
CREATE TYPE thread_type_enum AS ENUM ('general', 'portal', 'topic');
ALTER TABLE threads ADD COLUMN thread_type thread_type_enum NOT NULL DEFAULT 'general';
