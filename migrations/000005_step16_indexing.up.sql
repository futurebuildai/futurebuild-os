-- Migration: Step 16 Indexing
-- Adding indexes for common query patterns on projects

CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
CREATE INDEX IF NOT EXISTS idx_projects_name ON projects(name);
