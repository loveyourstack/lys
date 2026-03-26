
-- from https://stackoverflow.com/questions/36041784/postgresql-compare-two-jsonb-objects
CREATE OR REPLACE FUNCTION lyspgmon.jsonb_diff_val(val1 JSONB, val2 JSONB)
RETURNS JSONB AS $$
DECLARE
  result JSONB;
  v RECORD;
BEGIN
  result = val1;
  FOR v IN SELECT * FROM jsonb_each(val2) LOOP
    IF result @> jsonb_build_object(v.key,v.value)
      THEN result = result - v.key;
    ELSIF result ? v.key THEN CONTINUE;
    ELSE
      result = result || jsonb_build_object(v.key,'null');
    END IF;
  END LOOP;
  RETURN result;
END;
$$ LANGUAGE plpgsql;