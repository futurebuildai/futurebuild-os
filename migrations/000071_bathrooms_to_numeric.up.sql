-- Change bathrooms column from INTEGER to NUMERIC to support half-baths (2.5, 3.5)
ALTER TABLE projects
  ALTER COLUMN bathrooms TYPE NUMERIC(4,1) USING bathrooms::NUMERIC(4,1);
