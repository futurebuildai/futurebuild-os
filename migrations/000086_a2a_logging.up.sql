-- Migration: A2A Execution Logging
-- See FRONTEND_SCOPE.md Section 15.1 — OS-to-Brain UI Bridge

-- Agent-to-Agent execution logs for Why-Trail visualization
CREATE TABLE a2a_execution_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    workflow_id     UUID,
    source_system   VARCHAR(50) NOT NULL,
    target_system   VARCHAR(50) NOT NULL,
    action_type     VARCHAR(100) NOT NULL,
    payload         JSONB,
    status          VARCHAR(20) DEFAULT 'pending',
    error_message   TEXT,
    duration_ms     INT,
    executed_at     TIMESTAMPTZ DEFAULT NOW(),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_a2a_execution_logs_org ON a2a_execution_logs(org_id);
CREATE INDEX idx_a2a_execution_logs_workflow ON a2a_execution_logs(workflow_id);
CREATE INDEX idx_a2a_execution_logs_executed ON a2a_execution_logs(executed_at);
CREATE INDEX idx_a2a_execution_logs_status ON a2a_execution_logs(status);

-- Active agent connections (OS-to-Brain bridge status)
CREATE TABLE active_agent_connections (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    agent_name          VARCHAR(100) NOT NULL,
    agent_type          VARCHAR(50) NOT NULL,
    brain_workflow_id   UUID,
    status              VARCHAR(20) DEFAULT 'active',
    last_execution_at   TIMESTAMPTZ,
    execution_count     INT DEFAULT 0,
    error_count         INT DEFAULT 0,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(org_id, agent_name)
);

CREATE INDEX idx_active_agent_connections_org ON active_agent_connections(org_id);
CREATE INDEX idx_active_agent_connections_status ON active_agent_connections(status);

CREATE TRIGGER update_active_agent_connections_modtime
    BEFORE UPDATE ON active_agent_connections
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
