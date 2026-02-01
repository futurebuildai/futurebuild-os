-- Rollback: Remove building specification columns from projects table.

ALTER TABLE projects
  DROP COLUMN IF EXISTS bedrooms,
  DROP COLUMN IF EXISTS bathrooms,
  DROP COLUMN IF EXISTS stories,
  DROP COLUMN IF EXISTS lot_size,
  DROP COLUMN IF EXISTS foundation_type,
  DROP COLUMN IF EXISTS topography,
  DROP COLUMN IF EXISTS soil_conditions;
