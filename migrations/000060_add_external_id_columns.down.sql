-- Step 80: Rollback external_id columns

DROP INDEX IF EXISTS idx_users_external_id;
DROP INDEX IF EXISTS idx_organizations_external_id;

ALTER TABLE users DROP COLUMN IF EXISTS external_id;
ALTER TABLE organizations DROP COLUMN IF EXISTS external_id;
