-- Expand contacts for CRM capabilities and portal access tiers.
-- See FRONTEND_V2_SPEC.md §13

-- User-editable fields
ALTER TABLE contacts ADD COLUMN trades TEXT[] DEFAULT '{}';
ALTER TABLE contacts ADD COLUMN license_number VARCHAR(100);
ALTER TABLE contacts ADD COLUMN address_city VARCHAR(100);
ALTER TABLE contacts ADD COLUMN address_state VARCHAR(50);
ALTER TABLE contacts ADD COLUMN address_zip VARCHAR(20);
ALTER TABLE contacts ADD COLUMN website VARCHAR(500);
ALTER TABLE contacts ADD COLUMN notes TEXT;
ALTER TABLE contacts ADD COLUMN portal_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE contacts ADD COLUMN source VARCHAR(50) DEFAULT 'manual';

-- Agent-computed fields
ALTER TABLE contacts ADD COLUMN last_contacted_at TIMESTAMPTZ;
ALTER TABLE contacts ADD COLUMN avg_response_time_hours NUMERIC(6,1);
ALTER TABLE contacts ADD COLUMN on_time_rate NUMERIC(4,2);
ALTER TABLE contacts ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();

-- Trade-based lookups (GIN index for array contains queries)
CREATE INDEX idx_contacts_trades ON contacts USING GIN(trades);

-- Portal-enabled contacts
CREATE INDEX idx_contacts_portal ON contacts(org_id) WHERE portal_enabled = TRUE;

-- Geographic proximity matching
CREATE INDEX idx_contacts_zip ON contacts(org_id, address_zip) WHERE address_zip IS NOT NULL;

-- Auto-update updated_at trigger
CREATE OR REPLACE FUNCTION update_contacts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER contacts_updated_at
    BEFORE UPDATE ON contacts
    FOR EACH ROW
    EXECUTE FUNCTION update_contacts_updated_at();
