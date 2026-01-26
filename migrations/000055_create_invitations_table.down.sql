-- Rollback: Drop invitations table.
-- See LAUNCH_STRATEGY.md Task B2: User Invite Flow.

DROP INDEX IF EXISTS idx_invitations_org_pending;
DROP INDEX IF EXISTS idx_invitations_token_hash;
DROP TABLE IF EXISTS invitations;
