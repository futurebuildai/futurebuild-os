-- Rollback staging test users
-- Only removes grant (colton is also managed by 000058)
DELETE FROM users WHERE email = 'grant@futurebuild.ai';
DELETE FROM business_config WHERE org_id = '00000000-0000-0000-0000-000000000001';
