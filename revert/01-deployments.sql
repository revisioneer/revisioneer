-- Revert 01-deployments

BEGIN;

DROP TABLE public.deployments;

COMMIT;
