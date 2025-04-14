
DROP VIEW IF EXISTS lyspgmon.v_queries;

CREATE OR REPLACE VIEW lyspgmon.v_queries AS
  SELECT
    application_name,
    client_addr::text,
    pid,
    query,
    query_start,
    state,
    usename
  FROM pg_stat_activity 
  WHERE state IS NOT NULL 
    AND query NOT LIKE '%lyspgmon.v_queries%' 
    AND datname = current_database();