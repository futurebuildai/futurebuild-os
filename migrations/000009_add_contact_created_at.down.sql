-- 000009_add_contact_created_at.down.sql
ALTER TABLE contacts DROP COLUMN IF EXISTS created_at;
