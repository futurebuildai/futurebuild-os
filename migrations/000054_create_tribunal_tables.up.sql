-- Shadow Viewer: Tribunal Tables
-- See specs/SHADOW_VIEWER_specs.md Section 4.1

-- Tribunal Decision Status Enum
CREATE TYPE tribunal_decision_status AS ENUM ('APPROVED', 'REJECTED', 'CONFLICT');

-- Tribunal Vote Type Enum
CREATE TYPE tribunal_vote_type AS ENUM ('YEA', 'NAY', 'ABSTAIN');

-- Tribunal Decisions Table
-- Stores the high-level outcome of a Tribunal decision
CREATE TABLE tribunal_decisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    case_id VARCHAR(255) NOT NULL,
    context_summary TEXT NOT NULL,
    status tribunal_decision_status NOT NULL,
    consensus_score FLOAT NOT NULL CHECK (consensus_score >= 0 AND consensus_score <= 1),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Indexes for filtering (spec Section 4.1)
CREATE INDEX idx_tribunal_decisions_status_created ON tribunal_decisions (status, created_at DESC);
CREATE INDEX idx_tribunal_decisions_case_id ON tribunal_decisions (case_id);

-- Tribunal Votes Table
-- Stores individual model votes for each decision
CREATE TABLE tribunal_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    decision_id UUID NOT NULL REFERENCES tribunal_decisions(id) ON DELETE CASCADE,
    model_name VARCHAR(255) NOT NULL,
    vote tribunal_vote_type NOT NULL,
    reasoning TEXT,
    latency_ms INT NOT NULL DEFAULT 0,
    token_count INT NOT NULL DEFAULT 0,
    cost_usd DECIMAL(10, 6) NOT NULL DEFAULT 0
);

CREATE INDEX idx_tribunal_votes_decision_id ON tribunal_votes (decision_id);
