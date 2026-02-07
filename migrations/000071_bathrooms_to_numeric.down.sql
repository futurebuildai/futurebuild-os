-- Revert bathrooms column to INTEGER (will truncate decimal values)
ALTER TABLE projects
  ALTER COLUMN bathrooms TYPE INTEGER USING bathrooms::INTEGER;
