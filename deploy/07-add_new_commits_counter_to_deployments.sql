-- Deploy 07-add_new_commits_counter_to_deployments
-- requires: 01-deployments

BEGIN;

SET client_min_messages = 'warning';

ALTER TABLE public.deployments
  ADD COLUMN new_commit_counter int
  DEFAULT 0;

COMMIT;
