-- Feed Cards: materialized feed table written by agents, read by portfolio endpoint.
-- See FRONTEND_V2_SPEC.md §6.2

CREATE TABLE feed_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    card_type TEXT NOT NULL,
    priority INT NOT NULL DEFAULT 2,
    headline TEXT NOT NULL,
    body TEXT NOT NULL,
    consequence TEXT,
    horizon TEXT NOT NULL CHECK (horizon IN ('today', 'this_week', 'horizon')),
    deadline TIMESTAMPTZ,
    actions JSONB NOT NULL DEFAULT '[]',
    engine_data JSONB,
    agent_source TEXT,
    task_id UUID REFERENCES project_tasks(id) ON DELETE SET NULL,
    dismissed_at TIMESTAMPTZ,
    snoozed_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

-- Primary query: active cards for an org, sorted by priority
CREATE INDEX idx_feed_cards_org_active ON feed_cards(org_id, priority, created_at DESC)
    WHERE dismissed_at IS NULL AND (snoozed_until IS NULL OR snoozed_until < NOW());

-- Filter by project
CREATE INDEX idx_feed_cards_project ON feed_cards(project_id, created_at DESC)
    WHERE dismissed_at IS NULL;

-- Deduplication: prevent duplicate cards for the same task+type
CREATE UNIQUE INDEX idx_feed_cards_dedup ON feed_cards(project_id, card_type, task_id)
    WHERE task_id IS NOT NULL AND dismissed_at IS NULL;

-- Expiry cleanup
CREATE INDEX idx_feed_cards_expires ON feed_cards(expires_at)
    WHERE expires_at IS NOT NULL;
