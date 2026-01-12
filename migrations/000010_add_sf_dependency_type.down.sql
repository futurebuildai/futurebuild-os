-- PostgreSQL does not support removing enum values directly.
-- This migration is not reversible without recreating the type.
-- See: https://www.postgresql.org/docs/current/sql-altertype.html

-- MANUAL ROLLBACK REQUIRED if needed.
SELECT 'SF enum cannot be removed; manual intervention required' AS warning;
