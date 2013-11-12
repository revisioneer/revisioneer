-- Revert foreign_keys_on_deployments

BEGIN;

ALTER TABLE public.deployments
  DROP COLUMN project_id;

COMMIT;
