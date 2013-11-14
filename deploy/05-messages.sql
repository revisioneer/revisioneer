-- Deploy 05-messages
-- requires: 01-deployments

BEGIN;

SET client_min_messages = 'warning';

CREATE TABLE public.messages (
  id            SERIAL      PRIMARY KEY,
  message       TEXT        NOT NULL,
  deployment_id INT         NOT NULL
);

COMMIT;
