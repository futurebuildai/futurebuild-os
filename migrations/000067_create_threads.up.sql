-- 000067_create_threads.up.sql
-- Add conversation threads to projects.
-- See Thread Support Implementation Plan.

BEGIN;

-- 1. Create threads table
CREATE TABLE IF NOT EXISTS threads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    title TEXT NOT NULL,
    is_general BOOLEAN NOT NULL DEFAULT false,
    archived_at TIMESTAMPTZ,
    created_by UUID REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Partial unique index: guarantees exactly one General thread per project
CREATE UNIQUE INDEX idx_threads_project_general ON threads (project_id) WHERE is_general = true;

-- Standard lookup index
CREATE INDEX idx_threads_project_id ON threads (project_id);

-- 2. Add thread_id to chat_messages (nullable for backfill)
ALTER TABLE chat_messages ADD COLUMN thread_id UUID;

-- 3. Backfill: insert a General thread for each project that has chat messages
INSERT INTO threads (id, project_id, title, is_general, created_at, updated_at)
SELECT gen_random_uuid(), p.id, 'General', true, now(), now()
FROM projects p
WHERE EXISTS (SELECT 1 FROM chat_messages cm WHERE cm.project_id = p.id)
ON CONFLICT DO NOTHING;

-- Also insert General threads for projects without messages (ensures all projects have one)
INSERT INTO threads (id, project_id, title, is_general, created_at, updated_at)
SELECT gen_random_uuid(), p.id, 'General', true, now(), now()
FROM projects p
WHERE NOT EXISTS (SELECT 1 FROM threads t WHERE t.project_id = p.id AND t.is_general = true)
ON CONFLICT DO NOTHING;

-- 4. Backfill: update existing chat_messages to point to their project's General thread
UPDATE chat_messages cm
SET thread_id = t.id
FROM threads t
WHERE t.project_id = cm.project_id AND t.is_general = true AND cm.thread_id IS NULL;

-- 5. Make thread_id NOT NULL after backfill
ALTER TABLE chat_messages ALTER COLUMN thread_id SET NOT NULL;

-- 6. Add FK constraint and index
ALTER TABLE chat_messages ADD CONSTRAINT fk_chat_messages_thread_id
    FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE RESTRICT;

CREATE INDEX idx_chat_messages_thread_id_created_at ON chat_messages (thread_id, created_at);

COMMIT;
