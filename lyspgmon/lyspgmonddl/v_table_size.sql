
DROP VIEW IF EXISTS lyspgmon.v_table_size;

-- based on https://wiki.postgresql.org/wiki/Disk_Usage

CREATE VIEW lyspgmon.v_table_size AS
  WITH RECURSIVE pg_inherit(inhrelid, inhparent) AS
      (SELECT inhrelid, inhparent
      FROM pg_inherits
      UNION
      SELECT child.inhrelid, parent.inhparent
      FROM pg_inherit child, pg_inherits parent
      WHERE child.inhparent = parent.inhrelid),
  pg_inherit_short AS (SELECT * FROM pg_inherit WHERE inhparent NOT IN (SELECT inhrelid FROM pg_inherit))
  SELECT 
      table_schema::text,
      TABLE_NAME::text,
      row_estimate,
      total_bytes,
      pg_size_pretty(total_bytes) AS total_pretty,
      index_bytes,
      pg_size_pretty(index_bytes) AS index_pretty,
      COALESCE(toast_bytes,0) AS toast_bytes,
      COALESCE(pg_size_pretty(toast_bytes),'') AS toast_pretty,
      table_bytes,
      pg_size_pretty(table_bytes) AS table_pretty,
      total_bytes::float8 / sum(total_bytes) OVER () AS total_size_share
    FROM (
      SELECT *, total_bytes-index_bytes-COALESCE(toast_bytes,0) AS table_bytes
      FROM (
          SELECT c.oid
                , nspname AS table_schema
                , relname AS TABLE_NAME
                , SUM(c.reltuples) OVER (partition BY parent) AS row_estimate
                , SUM(pg_total_relation_size(c.oid)) OVER (partition BY parent) AS total_bytes
                , SUM(pg_indexes_size(c.oid)) OVER (partition BY parent) AS index_bytes
                , SUM(pg_total_relation_size(reltoastrelid)) OVER (partition BY parent) AS toast_bytes
                , parent
            FROM (
                  SELECT pg_class.oid
                      , reltuples
                      , relname
                      , relnamespace
                      , pg_class.reltoastrelid
                      , COALESCE(inhparent, pg_class.oid) parent
                  FROM pg_class
                  LEFT JOIN pg_inherit_short ON inhrelid = oid
                  WHERE relkind IN ('r', 'p')
              ) c
              LEFT JOIN pg_namespace n ON n.oid = c.relnamespace
    ) a
    WHERE oid = parent
  ) a
  WHERE table_schema::text NOT IN ('pg_catalog', 'information_schema');
