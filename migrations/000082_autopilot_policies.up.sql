CREATE TABLE autopilot_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    action_type TEXT NOT NULL,
    auto_approve BOOLEAN DEFAULT false,
    max_cost_cents BIGINT DEFAULT 0,
    require_approval_from TEXT[] DEFAULT '{}',
    cooldown_minutes INT DEFAULT 0,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(org_id, action_type)
);

CREATE INDEX idx_autopilot_policies_org ON autopilot_policies(org_id);

CREATE TRIGGER update_autopilot_policies_modtime
    BEFORE UPDATE ON autopilot_policies
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
