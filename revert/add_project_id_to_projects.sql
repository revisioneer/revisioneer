-- Revert add_project_id_to_projects

BEGIN;

ALTER TABLE public.deployments
  DROP COLUMN project_id;

COMMIT;
