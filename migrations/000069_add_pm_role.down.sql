-- PostgreSQL does not support removing enum values.
-- No-op: the PM value will remain in the enum but won't be assigned to new users.
SELECT 1;
