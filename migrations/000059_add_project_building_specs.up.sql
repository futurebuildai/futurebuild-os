-- Phase 11: Add building specification columns to projects table.
-- These fields are collected during onboarding and used by the physics engine (DHSM).
-- See STEP_74_SPLIT_SCREEN_WIZARD.md, STEP_76_REALTIME_FORM_FILLING.md

ALTER TABLE projects
  ADD COLUMN IF NOT EXISTS bedrooms INTEGER DEFAULT 0,
  ADD COLUMN IF NOT EXISTS bathrooms INTEGER DEFAULT 0,
  ADD COLUMN IF NOT EXISTS stories INTEGER DEFAULT 1,
  ADD COLUMN IF NOT EXISTS lot_size FLOAT DEFAULT 0.0,
  ADD COLUMN IF NOT EXISTS foundation_type VARCHAR(50) DEFAULT '',
  ADD COLUMN IF NOT EXISTS topography VARCHAR(50) DEFAULT '',
  ADD COLUMN IF NOT EXISTS soil_conditions VARCHAR(50) DEFAULT '';
