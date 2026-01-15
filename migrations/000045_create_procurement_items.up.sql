-- See BACKEND_SCOPE.md Section 4.2 (Procurement)
-- See PRODUCTION_PLAN.md Step 46

CREATE TABLE procurement_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_task_id UUID NOT NULL UNIQUE REFERENCES project_tasks(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    lead_time_weeks INT NOT NULL DEFAULT 4,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    calculated_order_date TIMESTAMPTZ,
    last_checked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_procurement_status CHECK (status IN ('pending', 'ok', 'warning', 'critical'))
);

CREATE INDEX idx_procurement_items_status ON procurement_items(status);
CREATE INDEX idx_procurement_items_project_task ON procurement_items(project_task_id);
