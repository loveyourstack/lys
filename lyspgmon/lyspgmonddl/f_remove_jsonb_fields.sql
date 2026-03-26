
CREATE OR REPLACE FUNCTION lyspgmon.remove_jsonb_fields(p_row jsonb)
  RETURNS jsonb AS
$BODY$
DECLARE
BEGIN

-- remove these cols in all tables
p_row = p_row - 'id';
p_row = p_row - 'created_at';
p_row = p_row - 'created_by';
p_row = p_row - 'last_user_update_by';
p_row = p_row - 'updated_at';

RETURN p_row;

END;
$BODY$
  LANGUAGE plpgsql VOLATILE
  COST 100;
