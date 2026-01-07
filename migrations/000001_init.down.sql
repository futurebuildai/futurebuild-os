-- Rollback: Initial Schema
-- Drops Domain 1: Identity & Access (IAM)

DROP TABLE IF EXISTS project_budgets;
DROP TABLE IF EXISTS project_assignments;
DROP TABLE IF EXISTS task_dependencies;
DROP TABLE IF EXISTS project_tasks;
DROP TABLE IF EXISTS wbs_tasks;
DROP TABLE IF EXISTS wbs_phases;
DROP TABLE IF EXISTS wbs_templates;
DROP TABLE IF EXISTS project_context;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS contacts;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS organizations;

DROP TYPE IF EXISTS dependency_type_enum;
DROP TYPE IF EXISTS task_status_type;
DROP TYPE IF EXISTS project_status_type;
DROP TYPE IF EXISTS contact_preference_type;
DROP TYPE IF EXISTS contact_role_type;
DROP TYPE IF EXISTS user_role_type;

-- Optional: Extensions are typically left in place.
-- DROP EXTENSION IF EXISTS "pgvector";
-- DROP EXTENSION IF EXISTS "uuid-ossp";
