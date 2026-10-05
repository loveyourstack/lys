
-- returns a table of dependent objects for the specified schema and object name
-- _name can be a table, view, materialized view, or SQL-standard function or procedure
CREATE OR REPLACE FUNCTION lyspgmon.f_dependents(_schema text, _name text)
RETURNS TABLE (
  "schema" text,
  name text,
  type text,
  depth int,
  drop_cmd text,
  create_cmt text
)
LANGUAGE SQL
BEGIN ATOMIC

/*
    note re: functions and procedures
    
    this function will only track dependencies for functions and procedures if they are written using SQL-standard form (BEGIN ATOMIC)
    if they are written with plpgsql, the content is created dynamically and cannot be tracked by this function
*/

WITH RECURSIVE target AS (
  SELECT 'pg_catalog.pg_class'::regclass AS classid, c.oid AS objid
  FROM pg_catalog.pg_class AS c
  JOIN pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
  WHERE n.nspname = _schema
  AND c.relname = _name
  AND c.relkind IN ('r', 'p', 'v', 'm', 'f') -- regular tables, partitioned tables, views, materialized views, foreign tables

  UNION ALL

  SELECT 'pg_catalog.pg_proc'::regclass, p.oid
  FROM pg_catalog.pg_proc AS p
  JOIN pg_catalog.pg_namespace AS n ON n.oid = p.pronamespace
  WHERE n.nspname = _schema
  AND p.proname = _name
  AND p.prokind IN ('f', 'w', 'p') -- may match multiple overloads; see note below
),
dep_edges AS (
  -- views / materialized views depending on a relation or a function
  SELECT d.refclassid, d.refobjid,
         'pg_catalog.pg_class'::regclass AS classid, rw.ev_class AS objid
  FROM pg_catalog.pg_depend AS d
  JOIN pg_catalog.pg_rewrite AS rw
    ON d.classid = 'pg_catalog.pg_rewrite'::regclass
    AND d.objid = rw.oid
  JOIN pg_catalog.pg_class AS dependent ON dependent.oid = rw.ev_class
  WHERE d.refclassid IN ('pg_catalog.pg_class'::regclass, 'pg_catalog.pg_proc'::regclass)
  AND dependent.relkind IN ('v', 'm')
  AND rw.ev_class <> d.refobjid -- exclude view self-reference

  UNION ALL

  -- SQL-standard functions depending on a relation or another function
  SELECT d.refclassid, d.refobjid,
         'pg_catalog.pg_proc'::regclass, p.oid
  FROM pg_catalog.pg_depend AS d
  JOIN pg_catalog.pg_proc AS p
    ON d.classid = 'pg_catalog.pg_proc'::regclass
    AND d.objid = p.oid
  WHERE d.refclassid IN ('pg_catalog.pg_class'::regclass, 'pg_catalog.pg_proc'::regclass)
  AND p.prokind IN ('f', 'w', 'p')
),
tree AS (
  SELECT t.classid, t.objid, 0 AS depth,
         ARRAY[t.classid::text || ':' || t.objid::text] AS path
  FROM target AS t

  UNION ALL

  SELECT e.classid, e.objid, tr.depth + 1,
         tr.path || (e.classid::text || ':' || e.objid::text)
  FROM tree AS tr
  JOIN dep_edges AS e
    ON e.refclassid = tr.classid
    AND e.refobjid = tr.objid
  WHERE NOT (e.classid::text || ':' || e.objid::text) = ANY(tr.path)
),
node_depth AS (
  SELECT tr.classid, tr.objid, max(tr.depth) AS depth
  FROM tree AS tr
  GROUP BY tr.classid, tr.objid
),
candidates AS (
  SELECT n.nspname AS schema_name, c.relname AS object_name,
         CASE c.relkind WHEN 'm' THEN 'materialized view' ELSE 'view' END AS object_type,
         nd.depth
  FROM node_depth AS nd
  JOIN pg_catalog.pg_class AS c ON nd.classid = 'pg_catalog.pg_class'::regclass AND c.oid = nd.objid
  JOIN pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
  WHERE NOT EXISTS (SELECT 1 FROM target t WHERE t.classid = nd.classid AND t.objid = nd.objid)

  UNION ALL

  SELECT n.nspname, p.proname, 'function', nd.depth
  FROM node_depth AS nd
  JOIN pg_catalog.pg_proc AS p ON nd.classid = 'pg_catalog.pg_proc'::regclass AND p.oid = nd.objid
  JOIN pg_catalog.pg_namespace AS n ON n.oid = p.pronamespace
  WHERE NOT EXISTS (SELECT 1 FROM target t WHERE t.classid = nd.classid AND t.objid = nd.objid)
)

SELECT 
  c.schema_name,
  c.object_name,
  c.object_type,
  min(c.depth)::int AS depth,
  CASE WHEN c.object_type = 'function' THEN 'DROP FUNCTION ' || c.schema_name || '.' || c.object_name || ';'
    WHEN c.object_type = 'materialized view' THEN 'DROP MATERIALIZED VIEW ' || c.schema_name || '.' || c.object_name || ';'
    WHEN c.object_type = 'view' THEN 'DROP VIEW ' || c.schema_name || '.' || c.object_name || ';'
    ELSE ''
  END AS drop_cmd,
  CASE WHEN c.object_type = 'function' THEN '-- + ' || c.schema_name || '.' || c.object_name || ';'
    WHEN c.object_type = 'materialized view' THEN '-- + ' || c.schema_name || '.' || c.object_name || ';'
    WHEN c.object_type = 'view' THEN '-- + ' || c.schema_name || '.' || c.object_name || ';'
    ELSE ''
  END AS create_cmt
FROM candidates AS c
GROUP BY c.schema_name, c.object_name, c.object_type

UNION ALL

-- no matching object: raise an error
SELECT NULL, NULL,
  lyspgmon.f_raise(format('lyspgmon.f_dependents: no table, view or SQL function found for %I.%I', _schema, _name)),
  NULL, NULL, NULL
WHERE NOT EXISTS (SELECT 1 FROM target);

END;