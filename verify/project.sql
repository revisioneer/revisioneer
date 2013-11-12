-- Verify project

BEGIN;

SELECT id, name, api_token, created_at
  FROM public.projects
  WHERE FALSE;

ROLLBACK;
