-- Deploy add_foreignkey_constraint
-- requires: foreign_keys_on_deployments

BEGIN;

ALTER TABLE public.deployments
  ADD CONSTRAINT projectfk FOREIGN KEY (project_id) REFERENCES projects (id) MATCH FULL;

COMMIT;
