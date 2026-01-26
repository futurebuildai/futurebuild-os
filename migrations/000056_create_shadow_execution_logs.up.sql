-- Shadow Execution Logs: Action Bridge
-- See specs/FUTURESHADE_AGENTS_SPEC.md Section 4 (Action Bridge)

-- Execution Status Enum
CREATE TYPE shadow_execution_status AS ENUM ('PENDING', 'RUNNING', 'COMPLETED', 'FAILED');

-- Shadow Execution Logs Table
-- Tracks skill executions triggered by Tribunal decisions
CREATE TABLE shadow_execution_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    decision_id UUID NOT NULL REFERENCES tribunal_decisions(id) ON DELETE CASCADE,
    skill_id TEXT NOT NULL,
    parameters JSONB NOT NULL DEFAULT '{}',
    status shadow_execution_status NOT NULL DEFAULT 'PENDING',
    result_summary TEXT,
    error_message TEXT,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for querying by decision (common lookup pattern)
CREATE INDEX idx_shadow_execution_logs_decision_id ON shadow_execution_logs(decision_id);

-- Partial index for active executions (PENDING/RUNNING) - used by worker to find work
CREATE INDEX idx_shadow_execution_logs_status ON shadow_execution_logs(status) WHERE status IN ('PENDING', 'RUNNING');
