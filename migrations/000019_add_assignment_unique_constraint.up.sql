-- Add unique constraint to prevent duplicate phase assignments per project.
-- Reference: BACKEND_SCOPE.md Line 514

ALTER TABLE project_assignments 
ADD CONSTRAINT unique_project_phase_assignment UNIQUE (project_id, wbs_phase_id);
