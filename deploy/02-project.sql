-- Deploy 02-project

BEGIN;

SET client_min_messages = 'warning';

CREATE TABLE public.projects(
  id         SERIAL      PRIMARY KEY,
  name       TEXT        NOT NULL,
  api_token  TEXT        NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMIT;
