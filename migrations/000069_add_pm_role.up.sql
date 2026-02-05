-- Add PM (Project Manager) role to user_role_type enum.
-- PM can read/write tasks, documents, chat, budget — but cannot create projects or change org settings.
ALTER TYPE user_role_type ADD VALUE IF NOT EXISTS 'PM';
