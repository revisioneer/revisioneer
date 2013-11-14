-- Deploy 03-add_project_id_to_projects
-- requires: 02-project
-- requires: 01-deployments

BEGIN;

SET client_min_messages = 'warning';

ALTER TABLE public.deployments
  ADD COLUMN project_id int;

COMMIT;
