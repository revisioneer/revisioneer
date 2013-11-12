-- Deploy foreign_keys_on_deployments
-- requires: project
-- requires: deployments

BEGIN;

ALTER TABLE public.deployments
  ADD COLUMN project_id int;

COMMIT;
