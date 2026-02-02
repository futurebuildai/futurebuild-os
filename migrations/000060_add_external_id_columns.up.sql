-- Step 80: Organization Manager — Add external_id columns for Clerk sync
-- See PHASE_12_PRD.md Section: Organization Manager

ALTER TABLE organizations ADD COLUMN IF NOT EXISTS external_id VARCHAR(255) UNIQUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS external_id VARCHAR(255) UNIQUE;

CREATE INDEX IF NOT EXISTS idx_organizations_external_id ON organizations(external_id);
CREATE INDEX IF NOT EXISTS idx_users_external_id ON users(external_id);
