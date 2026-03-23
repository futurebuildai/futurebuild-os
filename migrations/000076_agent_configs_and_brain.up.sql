-- Agent configuration per org (JSONB for flexible per-agent settings)
CREATE TABLE IF NOT EXISTS agent_configs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    config      JSONB NOT NULL DEFAULT '{}',
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by  UUID,
    UNIQUE (org_id)
);

CREATE INDEX idx_agent_configs_org ON agent_configs(org_id);

-- FB-Brain connection configuration per org
CREATE TABLE IF NOT EXISTS brain_connections (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    brain_url       TEXT NOT NULL DEFAULT '',
    integration_key TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'disconnected',
    last_sync_at    TIMESTAMPTZ,
    platforms       JSONB NOT NULL DEFAULT '[]',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (org_id)
);

CREATE INDEX idx_brain_connections_org ON brain_connections(org_id);
