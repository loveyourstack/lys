
DROP VIEW IF EXISTS lyspgmon.v_duplicate_shortnames;

CREATE VIEW lyspgmon.v_duplicate_shortnames AS

  WITH comments AS (
    SELECT pg_catalog.obj_description(to_regclass(t.table_schema || '.' || t.table_name)) AS com, table_schema, table_name
    FROM information_schema.tables t 
    WHERE t.table_schema NOT IN ('pg_catalog', 'information_schema') AND t.table_type = 'BASE TABLE'
  ), dups AS (
    SELECT com, count(*)
    FROM comments
    WHERE com LIKE 'shortname:\ %'
    GROUP BY 1
    HAVING count(*) > 1
  )
  SELECT dups.com, table_schema, table_name
  FROM comments
  JOIN dups ON comments.com = dups.com;