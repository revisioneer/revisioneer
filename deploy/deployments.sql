-- Deploy deployments
-- requires: appschema

BEGIN;

SET client_min_messages = 'warning';

CREATE TABLE revisioneer.deployments (
  id          SERIAL         PRIMARY KEY,
  sha         TEXT           NOT NULL,
  deployed_at TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

COMMIT;
