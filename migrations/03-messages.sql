-- +migrate Up

SET client_min_messages = 'warning';

CREATE TABLE public.messages (
  id            SERIAL      PRIMARY KEY,
  message       TEXT        NOT NULL,
  deployment_id INT         NOT NULL,
  CONSTRAINT deploymentfk FOREIGN KEY (deployment_id) REFERENCES deployments (id) MATCH FULL
);

-- +migrate Down

DROP TABLE public.messages;