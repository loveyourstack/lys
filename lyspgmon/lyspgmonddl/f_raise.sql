
-- raises an exception with the supplied message
-- allows SQL-standard (BEGIN ATOMIC) functions, which cannot use RAISE themselves, to report errors
CREATE OR REPLACE FUNCTION lyspgmon.f_raise(_msg text)
RETURNS text
LANGUAGE plpgsql
AS $$
BEGIN
  RAISE EXCEPTION '%', _msg;
END;
$$;