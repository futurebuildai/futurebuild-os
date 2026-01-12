-- Migration: Add SF (Start-to-Finish) to dependency_type_enum
-- Required for frontend/backend parity per FRONTEND_TYPES_SPEC.md

ALTER TYPE dependency_type_enum ADD VALUE 'SF';
