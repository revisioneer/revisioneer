CREATE TABLE public.deployments (
  id                 SERIAL      PRIMARY KEY,
  sha                TEXT        NOT NULL,
  deployed_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  project_id         int         NOT NULL,
  new_commit_counter int         NOT NULL DEFAULT 0,
  verified           boolean     NOT NULL DEFAULT false,
  verified_at        TIMESTAMPTZ DEFAULT NULL,

  CONSTRAINT projectfk FOREIGN KEY (project_id) REFERENCES projects (id) MATCH FULL
);
