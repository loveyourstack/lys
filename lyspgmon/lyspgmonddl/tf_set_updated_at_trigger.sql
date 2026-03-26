
CREATE OR REPLACE FUNCTION lyspgmon.set_updated_at()
  RETURNS trigger AS
$BODY$
BEGIN

  NEW.updated_at := current_timestamp;
  RETURN NEW;

END;
$BODY$
LANGUAGE plpgsql VOLATILE
COST 100;
