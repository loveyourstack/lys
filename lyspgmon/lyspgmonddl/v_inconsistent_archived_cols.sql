
DROP VIEW IF EXISTS lyspgmon.v_inconsistent_archived_cols;

CREATE VIEW lyspgmon.v_inconsistent_archived_cols AS

  WITH base AS (
    SELECT c.table_schema, c.table_name, c.column_name
    FROM information_schema.columns c 
    JOIN information_schema.tables t USING (table_schema, table_name)
    WHERE c.table_schema NOT IN ('pg_catalog', 'information_schema') AND t.table_type = 'BASE TABLE' AND c.table_name NOT LIKE '%\_archived'
    AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = t.table_schema AND table_name = t.table_name || '_archived')
    ORDER BY 1,2,3
  ), arc AS (
    SELECT c.table_schema, c.table_name, c.column_name
    FROM information_schema.columns c 
    JOIN information_schema.tables t USING (table_schema, table_name)
    WHERE c.table_schema NOT IN ('pg_catalog', 'information_schema') AND t.table_type = 'BASE TABLE' AND c.table_name LIKE '%\_archived'
    AND c.column_name NOT LIKE 'archived\_%'
    ORDER BY 1,2,3
  )
  SELECT 'In base, missing in _archived' AS info, table_schema, table_name, column_name
  FROM base
  WHERE NOT EXISTS (SELECT 1 FROM arc WHERE table_schema = base.table_schema AND table_name = base.table_name || '_archived' AND column_name = base.column_name)
  
  UNION ALL
  
  SELECT 'In _archived, missing in base', table_schema, table_name, column_name
  FROM arc
  WHERE NOT EXISTS (SELECT 1 FROM base WHERE table_schema = arc.table_schema AND table_name || '_archived' = arc.table_name AND column_name = arc.column_name);
