-- Verify 07-add_new_commits_counter_to_deployments

BEGIN;

SELECT new_commit_counter
  FROM public.deployments
  WHERE false;

ROLLBACK;
