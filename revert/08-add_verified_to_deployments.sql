-- Revert 08-add_verified_to_deployments

BEGIN;

ALTER TABLE public.deployments
  DROP COLUMN verified;


COMMIT;
