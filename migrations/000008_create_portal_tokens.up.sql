-- 000008_create_portal_tokens.up.sql
-- Enable magic link authentication for CONTACTS (External Users)

CREATE TABLE IF NOT EXISTS portal_tokens (
    token_hash VARCHAR(64) PRIMARY KEY,
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    used BOOLEAN DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_portal_tokens_contact_id ON portal_tokens(contact_id);
