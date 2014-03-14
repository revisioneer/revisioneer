-- Verify 08-add_verified_to_deployments

BEGIN;

SELECT verified
  FROM public.deployments
  WHERE false;

ROLLBACK;
