-- Rollback: Restore original procurement status constraint

ALTER TABLE procurement_items DROP CONSTRAINT IF EXISTS chk_procurement_status;

ALTER TABLE procurement_items ADD CONSTRAINT chk_procurement_status 
    CHECK (status IN ('pending', 'ok', 'warning', 'critical'));
