-- Revert add_foreignkey_constraint

BEGIN;

ALTER TABLE public.deployments
  DROP CONSTRAINT projectfk;

COMMIT;
