DROP TRIGGER IF EXISTS contacts_updated_at ON contacts;
DROP FUNCTION IF EXISTS update_contacts_updated_at();

DROP INDEX IF EXISTS idx_contacts_zip;
DROP INDEX IF EXISTS idx_contacts_portal;
DROP INDEX IF EXISTS idx_contacts_trades;

ALTER TABLE contacts DROP COLUMN IF EXISTS updated_at;
ALTER TABLE contacts DROP COLUMN IF EXISTS on_time_rate;
ALTER TABLE contacts DROP COLUMN IF EXISTS avg_response_time_hours;
ALTER TABLE contacts DROP COLUMN IF EXISTS last_contacted_at;
ALTER TABLE contacts DROP COLUMN IF EXISTS source;
ALTER TABLE contacts DROP COLUMN IF EXISTS portal_enabled;
ALTER TABLE contacts DROP COLUMN IF EXISTS notes;
ALTER TABLE contacts DROP COLUMN IF EXISTS website;
ALTER TABLE contacts DROP COLUMN IF EXISTS address_zip;
ALTER TABLE contacts DROP COLUMN IF EXISTS address_state;
ALTER TABLE contacts DROP COLUMN IF EXISTS address_city;
ALTER TABLE contacts DROP COLUMN IF EXISTS license_number;
ALTER TABLE contacts DROP COLUMN IF EXISTS trades;
