
DROP VIEW IF EXISTS lyspgmon.v_missing_updated_at_trigger;

CREATE VIEW lyspgmon.v_missing_updated_at_trigger AS

  WITH needs AS (
    SELECT c.table_schema, c.table_name 
    FROM information_schema.columns c 
    JOIN information_schema.tables t USING (table_schema, table_name)
    WHERE c.table_schema NOT IN ('pg_catalog', 'information_schema') AND t.table_type = 'BASE TABLE' AND column_name = 'updated_at'
    AND c.table_name NOT LIKE '%\_archived'
    ORDER BY 1,2
  ), has_u AS (
    SELECT event_object_schema, event_object_table
    FROM information_schema.triggers 
    WHERE event_manipulation = 'UPDATE' AND trigger_name LIKE '%set\_updated\_at\_u'
  ), has_i AS (
    SELECT event_object_schema, event_object_table
    FROM information_schema.triggers 
    WHERE event_manipulation = 'INSERT' AND trigger_name LIKE '%set\_updated\_at\_i'
  )
  SELECT table_schema, table_name, 'UPDATE' AS event
  FROM needs
  WHERE NOT EXISTS (SELECT 1 FROM has_u WHERE table_schema = has_u.event_object_schema AND table_name = has_u.event_object_table)
  
  UNION ALL
  
  SELECT table_schema, table_name, 'INSERT'
  FROM needs
  WHERE NOT EXISTS (SELECT 1 FROM has_i WHERE table_schema = has_i.event_object_schema AND table_name = has_i.event_object_table);