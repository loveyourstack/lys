
DROP VIEW IF EXISTS lyspgmon.v_missing_last_user_update_by_col;

CREATE VIEW lyspgmon.v_missing_last_user_update_by_col AS

  WITH needs AS (
    SELECT event_object_schema, event_object_table
    FROM information_schema.triggers 
    WHERE event_manipulation = 'UPDATE' AND trigger_name = 't_audit_update'
    ORDER BY 1,2
  ), has AS (
    SELECT c.table_schema, c.table_name 
    FROM information_schema.columns c 
    JOIN information_schema.tables t USING (table_schema, table_name)
    WHERE c.table_schema NOT IN ('pg_catalog', 'information_schema') AND t.table_type = 'BASE TABLE' AND column_name = 'last_user_update_by'
    AND c.table_name NOT LIKE '%\_archived'
  )
  SELECT event_object_schema, event_object_table
  FROM needs
  WHERE NOT EXISTS (SELECT 1 FROM has WHERE event_object_schema = has.table_schema AND event_object_table = has.table_name);