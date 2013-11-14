-- Verify 06-add_messages_deployment_fk

BEGIN;

CREATE OR REPLACE FUNCTION test_constraint(text, text)
  RETURNS void LANGUAGE plpgsql AS
$BODY$
BEGIN

IF (SELECT (SELECT count(*)
  FROM information_schema.table_constraints
  WHERE table_name=$1
    AND constraint_name=$2) != 1)
THEN
   RAISE EXCEPTION '% in % is missing', $2, $1;
END IF;

END;
$BODY$;

SELECT test_constraint('messages', 'deploymentfk');

ROLLBACK;
