-- Create documents table if it doesn't exist (Backfill from Phase 5, Step 35 requirement)
CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    type VARCHAR(50),
    filename VARCHAR(255),
    storage_path TEXT,
    mime_type VARCHAR(100),
    file_size_bytes BIGINT,
    processing_status VARCHAR(50) DEFAULT 'pending',
    extracted_text TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    uploaded_by UUID REFERENCES users(id) ON DELETE SET NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for project-based lookups
CREATE INDEX IF NOT EXISTS idx_documents_project_id ON documents(project_id);
