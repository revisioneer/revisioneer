-- Revert 06-add_messages_deployment_fk

BEGIN;

ALTER TABLE public.messages
  DROP CONSTRAINT deploymentfk;

COMMIT;
