-- Verify deployments

BEGIN;

SELECT id, sha, deployed_at
  FROM revisioneer.deployments
  WHERE FALSE;

ROLLBACK;
