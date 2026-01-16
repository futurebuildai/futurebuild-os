-- Migration: Add config_error status to procurement_items
-- Reason: ConfigurationError status is needed when location data is missing
-- See "Procurement Agent Data Accuracy" remediation (FAANG: Fail Loudly)

-- Drop old constraint
ALTER TABLE procurement_items DROP CONSTRAINT IF EXISTS chk_procurement_status;

-- Add new constraint with config_error status
ALTER TABLE procurement_items ADD CONSTRAINT chk_procurement_status 
    CHECK (status IN ('pending', 'ok', 'warning', 'critical', 'config_error'));
