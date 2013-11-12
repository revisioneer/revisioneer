-- Revert deployments

BEGIN;

DROP TABLE public.deployments;

COMMIT;
