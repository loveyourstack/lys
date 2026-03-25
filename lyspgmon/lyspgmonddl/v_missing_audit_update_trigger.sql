
DROP VIEW IF EXISTS lyspgmon.v_missing_audit_update_trigger;

CREATE VIEW lyspgmon.v_missing_audit_update_trigger AS

  WITH needs AS (
    SELECT v.table_schema, replace(replace(v.table_name, '_data_update', ''), 'v_', '') AS table_name
    FROM information_schema.views v
    WHERE v.table_schema NOT IN ('pg_catalog', 'information_schema', 'lyspgmon') 
    AND v.table_name NOT LIKE '%\_archived'
    AND v.table_name LIKE '%\_data_update'
    ORDER BY 1,2
  ), has AS (
    SELECT event_object_schema, event_object_table
    FROM information_schema.triggers 
    WHERE event_manipulation = 'UPDATE' AND trigger_name = 't_audit_update'
  )
  SELECT table_schema, table_name
  FROM needs
  WHERE NOT EXISTS (SELECT 1 FROM has WHERE table_schema = has.event_object_schema AND table_name = has.event_object_table);