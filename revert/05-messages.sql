-- Revert 05-messages

BEGIN;

DROP TABLE public.messages;

COMMIT;
