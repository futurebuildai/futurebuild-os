-- Seed staging test users for FutureBuild
-- Idempotent: uses ON CONFLICT to skip existing records
-- See Phase 15 staging readiness audit
--
-- Users:
--   colton@futurebuild.ai (already seeded in 000058, reinforced here)
--   grant@futurebuild.ai  (new)

-- Ensure the FutureBuild organization exists
INSERT INTO organizations (id, name, slug, project_limit)
VALUES ('00000000-0000-0000-0000-000000000001', 'FutureBuild', 'futurebuild', 25)
ON CONFLICT (slug) DO UPDATE SET project_limit = EXCLUDED.project_limit;

-- Seed colton (reinforces 000058, adds external_id placeholder)
INSERT INTO users (id, org_id, email, name, role)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000001',
    'colton@futurebuild.ai',
    'Colton',
    'Admin'
)
ON CONFLICT (email) DO NOTHING;

-- Seed grant
INSERT INTO users (id, org_id, email, name, role)
VALUES (
    '00000000-0000-0000-0000-000000000003',
    '00000000-0000-0000-0000-000000000001',
    'grant@futurebuild.ai',
    'Grant',
    'Admin'
)
ON CONFLICT (email) DO NOTHING;

-- Create a default business_config for the org so settings page works
INSERT INTO business_config (id, org_id, speed_multiplier, work_days)
VALUES (
    '00000000-0000-0000-0000-000000000010',
    '00000000-0000-0000-0000-000000000001',
    1.00,
    '[1, 2, 3, 4, 5]'::jsonb
)
ON CONFLICT (org_id) DO NOTHING;
