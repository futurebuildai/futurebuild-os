-- Agent Pending Actions: human-in-the-loop approval system for agent-recommended actions.
-- When an agent calls create_approval_card, the action payload is stored here.
-- The user approves/rejects via the feed card, and the action is executed or dismissed.
-- See plan Phase 5: Human-in-the-Loop Approval System.

CREATE TABLE agent_pending_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    feed_card_id UUID NOT NULL REFERENCES feed_cards(id) ON DELETE CASCADE,
    agent_source TEXT NOT NULL,
    action_type TEXT NOT NULL,
    action_payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'expired')),
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    rejection_reason TEXT,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Primary query: pending actions for an org
CREATE INDEX idx_agent_pending_org_status ON agent_pending_actions(org_id, status, created_at DESC)
    WHERE status = 'pending';

-- Lookup by feed card (1:1 relationship)
CREATE UNIQUE INDEX idx_agent_pending_card ON agent_pending_actions(feed_card_id);

-- Expiry cleanup
CREATE INDEX idx_agent_pending_expires ON agent_pending_actions(expires_at)
    WHERE status = 'pending' AND expires_at IS NOT NULL;
