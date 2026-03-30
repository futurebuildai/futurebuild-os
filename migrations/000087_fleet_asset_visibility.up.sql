-- Phase 20: Fleet asset visibility control
-- NULL or empty = visible to all roles; populated = only listed roles can see
ALTER TABLE fleet_assets ADD COLUMN visible_to_roles TEXT[] DEFAULT NULL;

CREATE INDEX idx_fleet_assets_visibility ON fleet_assets USING GIN (visible_to_roles);
