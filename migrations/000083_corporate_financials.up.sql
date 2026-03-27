-- Migration: Corporate Financials
-- See BACKEND_SCOPE.md Section 20.1 — Cross-project budget rollups, GL sync, AR aging

-- Corporate budget rollups (aggregates project_budgets by org)
CREATE TABLE corporate_budgets (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                 UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    fiscal_year            INT NOT NULL,
    quarter                INT NOT NULL CHECK (quarter BETWEEN 1 AND 4),
    total_estimated_cents  BIGINT NOT NULL DEFAULT 0,
    total_committed_cents  BIGINT NOT NULL DEFAULT 0,
    total_actual_cents     BIGINT NOT NULL DEFAULT 0,
    project_count          INT NOT NULL DEFAULT 0,
    last_rollup_at         TIMESTAMPTZ DEFAULT NOW(),
    created_at             TIMESTAMPTZ DEFAULT NOW(),
    updated_at             TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(org_id, fiscal_year, quarter)
);

CREATE INDEX idx_corporate_budgets_org ON corporate_budgets(org_id);
CREATE INDEX idx_corporate_budgets_year ON corporate_budgets(fiscal_year);

CREATE TRIGGER update_corporate_budgets_modtime
    BEFORE UPDATE ON corporate_budgets
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- GL sync log for QuickBooks/Xero/manual exports (audit trail)
CREATE TABLE gl_sync_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    sync_type       VARCHAR(50) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    records_synced  INT DEFAULT 0,
    error_message   TEXT,
    synced_at       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_gl_sync_logs_org ON gl_sync_logs(org_id);
CREATE INDEX idx_gl_sync_logs_status ON gl_sync_logs(status);

-- AP/AR aging snapshots for cash flow analysis
CREATE TABLE ar_aging_snapshots (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    snapshot_date           DATE NOT NULL,
    current_cents           BIGINT NOT NULL DEFAULT 0,
    days_30_cents           BIGINT NOT NULL DEFAULT 0,
    days_60_cents           BIGINT NOT NULL DEFAULT 0,
    days_90_plus_cents      BIGINT NOT NULL DEFAULT 0,
    total_receivable_cents  BIGINT NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(org_id, snapshot_date)
);

CREATE INDEX idx_ar_aging_snapshots_org ON ar_aging_snapshots(org_id);
CREATE INDEX idx_ar_aging_snapshots_date ON ar_aging_snapshots(snapshot_date);
