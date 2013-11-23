-- Revert 07-add_new_commits_counter_to_deployments

BEGIN;

ALTER TABLE public.deployments
  DROP COLUMN new_commit_counter;

COMMIT;
