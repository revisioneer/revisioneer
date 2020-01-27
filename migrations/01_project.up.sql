CREATE TABLE public.projects(
  id         SERIAL      PRIMARY KEY,
  name       TEXT        NOT NULL,
  api_token  TEXT        NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT uniq_api_token UNIQUE(api_token)
);
