-- Deploy 06-add_messages_deployment_fk
-- requires: 05-messages

BEGIN;

SET client_min_messages = 'warning';

ALTER TABLE public.messages
  ADD CONSTRAINT deploymentfk FOREIGN KEY (deployment_id) REFERENCES deployments (id) MATCH FULL;

COMMIT;
