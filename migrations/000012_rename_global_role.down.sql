-- Rollback: Rename role back to global_role

ALTER TABLE contacts RENAME COLUMN role TO global_role;
