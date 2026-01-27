-- Seed admin user for staging/demo environments
-- Uses ON CONFLICT to be idempotent (safe to run multiple times)

INSERT INTO organizations (id, name, slug)
VALUES ('00000000-0000-0000-0000-000000000001', 'FutureBuild', 'futurebuild')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO users (id, org_id, email, name, role)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000001',
    'colton@futurebuild.ai',
    'Colton',
    'Admin'
)
ON CONFLICT (email) DO NOTHING;
