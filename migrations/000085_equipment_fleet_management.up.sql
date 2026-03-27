-- Migration: Equipment & Fleet Asset Management (EAM)
-- See BACKEND_SCOPE.md Section 20.3 — Resource-constrained scheduling

-- Required for EXCLUDE USING GIST constraint on equipment_allocations
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- Fleet assets (heavy equipment, vehicles, tools)
CREATE TABLE fleet_assets (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    asset_number        VARCHAR(50) NOT NULL,
    asset_type          VARCHAR(100) NOT NULL,
    make                VARCHAR(100),
    model               VARCHAR(100),
    year                INT,
    vin                 VARCHAR(50),
    license_plate       VARCHAR(20),
    purchase_date       DATE,
    purchase_cost_cents BIGINT,
    current_value_cents BIGINT,
    status              VARCHAR(20) DEFAULT 'available',
    location            VARCHAR(255),
    notes               TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(org_id, asset_number)
);

CREATE INDEX idx_fleet_assets_org ON fleet_assets(org_id);
CREATE INDEX idx_fleet_assets_type ON fleet_assets(asset_type);
CREATE INDEX idx_fleet_assets_status ON fleet_assets(status);

CREATE TRIGGER update_fleet_assets_modtime
    BEFORE UPDATE ON fleet_assets
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- Equipment allocations (resource-constrained scheduling)
-- EXCLUDE USING GIST prevents double-booking: same asset cannot overlap dates
CREATE TABLE equipment_allocations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id        UUID NOT NULL REFERENCES fleet_assets(id) ON DELETE CASCADE,
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    task_id         UUID REFERENCES project_tasks(id) ON DELETE SET NULL,
    allocated_from  DATE NOT NULL,
    allocated_to    DATE NOT NULL,
    status          VARCHAR(20) DEFAULT 'planned',
    notes           TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT chk_allocation_dates CHECK (allocated_to >= allocated_from),
    CONSTRAINT no_overlap_allocation EXCLUDE USING GIST (
        asset_id WITH =,
        daterange(allocated_from, allocated_to, '[]') WITH &&
    ) WHERE (status IN ('planned', 'active'))
);

CREATE INDEX idx_equipment_allocations_asset ON equipment_allocations(asset_id);
CREATE INDEX idx_equipment_allocations_project ON equipment_allocations(project_id);
CREATE INDEX idx_equipment_allocations_task ON equipment_allocations(task_id);
CREATE INDEX idx_equipment_allocations_dates ON equipment_allocations(allocated_from, allocated_to);

CREATE TRIGGER update_equipment_allocations_modtime
    BEFORE UPDATE ON equipment_allocations
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- Maintenance logs for EAM
CREATE TABLE maintenance_logs (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id          UUID NOT NULL REFERENCES fleet_assets(id) ON DELETE CASCADE,
    maintenance_type  VARCHAR(50) NOT NULL,
    description       TEXT NOT NULL,
    scheduled_date    DATE,
    completed_date    DATE,
    cost_cents        BIGINT,
    vendor_name       VARCHAR(255),
    notes             TEXT,
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_maintenance_logs_asset ON maintenance_logs(asset_id);
CREATE INDEX idx_maintenance_logs_scheduled ON maintenance_logs(scheduled_date);

CREATE TRIGGER update_maintenance_logs_modtime
    BEFORE UPDATE ON maintenance_logs
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
