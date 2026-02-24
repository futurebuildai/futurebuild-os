CREATE TABLE IF NOT EXISTS audit_decisions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    agent VARCHAR(50) NOT NULL,
    action VARCHAR(100) NOT NULL,
    input_summary TEXT,
    decision TEXT,
    confidence DECIMAL(5,4),
    model VARCHAR(100),
    latency_ms BIGINT,
    project_id VARCHAR(255),
    user_id VARCHAR(255),
    trace_id VARCHAR(100),
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_decisions_project ON audit_decisions(project_id);
CREATE INDEX IF NOT EXISTS idx_audit_decisions_trace ON audit_decisions(trace_id);
CREATE INDEX IF NOT EXISTS idx_audit_decisions_timestamp ON audit_decisions(timestamp);
