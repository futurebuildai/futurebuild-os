-- Step 81 L7 Remediation: Add Viewer role to user_role_type enum.
-- See STEP_81_ROLE_MAPPING.md: Viewer = read-only access.
ALTER TYPE user_role_type ADD VALUE IF NOT EXISTS 'Viewer';
