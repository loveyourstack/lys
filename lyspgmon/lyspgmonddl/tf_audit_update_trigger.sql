
CREATE OR REPLACE FUNCTION lyspgmon.audit_update_trigger()
  RETURNS trigger AS
$BODY$
DECLARE
	v_old_row jsonb;
	v_new_row jsonb;
	v_old_values jsonb;
  v_new_values jsonb;
	v_count int;
  v_id bigint;
	v_user text;
BEGIN

IF TG_OP != 'UPDATE' THEN
  RETURN NEW;
END IF;

-- get old row details
v_old_row := to_jsonb(OLD);
v_id := v_old_row->'id';

-- remove old row columns that shouldn't be saved
v_old_row := lyspgmon.remove_jsonb_fields(v_old_row);

-- get new row details and remove unwanted fields
v_new_row := to_jsonb(NEW);
v_user := COALESCE(v_new_row->>'last_user_update_by', 'Unknown');

-- remove new row columns that shouldn't be saved
v_new_row := lyspgmon.remove_jsonb_fields(v_new_row);

-- evaluate difference between new and old rows
v_new_values := lyspgmon.jsonb_diff_val(v_new_row, v_old_row);

-- if no diff, exit
SELECT count(*) FROM jsonb_object_keys(v_new_values) INTO v_count;
IF v_count = 0 THEN
  RETURN NEW;
END IF;

-- evaluate difference between old and new rows
v_old_values := lyspgmon.jsonb_diff_val(v_old_row, v_new_row);

-- record change
INSERT INTO lyspgmon.audit_update (affected_schema, affected_table, affected_old_values, affected_new_values, affected_id, affected_by)
  VALUES (TG_TABLE_SCHEMA, TG_RELNAME, v_old_values, v_new_values, v_id, v_user);
RETURN NEW;

END;
$BODY$
LANGUAGE plpgsql VOLATILE
COST 100;
