-- Verify 05-messages

BEGIN;

SELECT id, message, deployment_id
  FROM public.messages
  WHERE FALSE;

ROLLBACK;
