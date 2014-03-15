-- Verify 09-add-verified_at_to_deployments

BEGIN;

SELECT verified_at
  FROM public.deployments
  WHERE false;

ROLLBACK;
