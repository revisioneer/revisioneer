-- Revert deployments

BEGIN;

DROP TABLE revisioneer.deployments;

COMMIT;
