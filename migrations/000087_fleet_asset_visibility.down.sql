DROP INDEX IF EXISTS idx_fleet_assets_visibility;
ALTER TABLE fleet_assets DROP COLUMN IF EXISTS visible_to_roles;
