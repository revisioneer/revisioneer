-- Deploy 08-add_verified_to_deployments
-- requires: 07-add_new_commits_counter_to_deployments

BEGIN;

SET client_min_messages = 'warning';

ALTER TABLE public.deployments
  ADD COLUMN verified boolean
  DEFAULT false;

COMMIT;
