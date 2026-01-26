-- Create invitations table for invite-only onboarding flow.
-- See LAUNCH_STRATEGY.md Task B2: User Invite Flow.

CREATE TABLE IF NOT EXISTS invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    role user_role_type NOT NULL DEFAULT 'Builder',
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure one active invitation per email per org
    CONSTRAINT invitations_unique_active UNIQUE (org_id, email, token_hash)
);

-- Index for token lookups during acceptance
CREATE INDEX idx_invitations_token_hash ON invitations(token_hash) WHERE accepted_at IS NULL;

-- Index for listing pending invitations by org
CREATE INDEX idx_invitations_org_pending ON invitations(org_id, created_at DESC) WHERE accepted_at IS NULL;

COMMENT ON TABLE invitations IS 'Pending user invitations for invite-only onboarding. See LAUNCH_STRATEGY.md';
COMMENT ON COLUMN invitations.token_hash IS 'SHA-256 hash of the invite token sent via email';
COMMENT ON COLUMN invitations.expires_at IS 'Invitation expiry time (default: 24 hours from creation)';
COMMENT ON COLUMN invitations.accepted_at IS 'When the invitation was accepted (NULL if pending)';
