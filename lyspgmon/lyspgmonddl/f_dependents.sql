
CREATE OR REPLACE FUNCTION lyspgmon.f_dependents(_schema text, _table text)
RETURNS TABLE (
  "schema" text,
  name text,
  type text,
  depth int
)
LANGUAGE SQL
BEGIN ATOMIC

WITH RECURSIVE target AS (
  SELECT c.oid AS relation_oid
  FROM pg_catalog.pg_class AS c
  JOIN pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
  WHERE n.nspname = _schema
  AND c.relname = _table
  AND c.relkind IN ('r', 'p', 'v', 'm', 'f') -- regular tables, partitioned tables, views, materialized views, foreign tables
),
view_dependencies AS (
  SELECT rw.ev_class AS dependent_oid, d.refobjid AS referenced_oid
  FROM pg_catalog.pg_depend AS d
  JOIN pg_catalog.pg_rewrite AS rw
    ON d.classid = 'pg_catalog.pg_rewrite'::regclass
    AND d.objid = rw.oid
  JOIN pg_catalog.pg_class AS dependent ON dependent.oid = rw.ev_class
  WHERE d.refclassid = 'pg_catalog.pg_class'::regclass
  AND dependent.relkind IN ('v', 'm')
  AND rw.ev_class <> d.refobjid
),
relation_tree AS (
  SELECT t.relation_oid, 0 AS depth, ARRAY[t.relation_oid] AS path
  FROM target AS t

  UNION ALL

  SELECT vd.dependent_oid, rt.depth + 1, rt.path || vd.dependent_oid
  FROM relation_tree AS rt
  JOIN view_dependencies AS vd ON vd.referenced_oid = rt.relation_oid
  WHERE NOT vd.dependent_oid = ANY(rt.path)
),
relation_depth AS (
  SELECT rt.relation_oid, max(rt.depth) AS depth
  FROM relation_tree AS rt
  GROUP BY rt.relation_oid
),
candidates AS (
  SELECT n.nspname AS schema_name,
    c.relname AS object_name,
  CASE c.relkind
    WHEN 'm' THEN 'materialized view'
    ELSE 'view'
  END AS object_type,
  rd.depth - 1 AS depth
  FROM relation_depth AS rd
  JOIN pg_catalog.pg_class AS c ON c.oid = rd.relation_oid
  JOIN pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
  JOIN target AS t ON t.relation_oid <> rd.relation_oid
  WHERE c.relkind IN ('v', 'm')

  UNION ALL

  SELECT n.nspname AS schema_name,
    p.proname AS object_name,
    'function' AS object_type,
  CASE
    WHEN rd.relation_oid = t.relation_oid THEN rd.depth
    ELSE rd.depth
  END AS depth
  FROM relation_depth AS rd
  JOIN target AS t ON true
  JOIN pg_catalog.pg_depend AS d
    ON d.refclassid = 'pg_catalog.pg_class'::regclass
    AND d.refobjid = rd.relation_oid
  JOIN pg_catalog.pg_proc AS p
    ON d.classid = 'pg_catalog.pg_proc'::regclass
    AND d.objid = p.oid
  JOIN pg_catalog.pg_namespace AS n ON n.oid = p.pronamespace
  WHERE p.prokind IN ('f', 'w') -- functions and window functions
)
SELECT c.schema_name, c.object_name, c.object_type, min(c.depth)::int AS depth
FROM candidates AS c
GROUP BY c.schema_name, c.object_name, c.object_type;

END;
