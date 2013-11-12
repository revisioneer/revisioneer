-- Verify 03-add_project_id_to_projects

BEGIN;

SELECT project_id
  FROM public.deployments
  WHERE false;

ROLLBACK;
