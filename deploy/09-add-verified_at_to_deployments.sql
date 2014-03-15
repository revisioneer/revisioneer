-- Deploy 09-add-verified_at_to_deployments
-- requires: 01-deployments

BEGIN;

SET client_min_messages = 'warning';

ALTER TABLE public.deployments
  ADD COLUMN verified_at TIMESTAMPTZ
  DEFAULT NULL;

COMMIT;
