-- Migration: Rename global_role to role for spec parity
-- Reference: DATA_SPINE_SPEC.md Section 2.3

ALTER TABLE contacts RENAME COLUMN global_role TO role;
