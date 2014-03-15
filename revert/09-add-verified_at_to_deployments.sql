-- Revert 09-add-verified_at_to_deployments

BEGIN;

ALTER TABLE public.deployments
  DROP COLUMN verified_at;

COMMIT;
