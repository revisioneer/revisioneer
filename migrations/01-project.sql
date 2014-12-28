-- +migrate Up

CREATE TABLE public.projects(
  id         SERIAL      PRIMARY KEY,
  name       TEXT        NOT NULL,
  api_token  TEXT        NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +migrate Down

DROP TABLE public.projects;