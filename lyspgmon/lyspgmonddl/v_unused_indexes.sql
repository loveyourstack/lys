
DROP VIEW IF EXISTS lyspgmon.v_unused_indexes;

-- based on https://github.com/dlamotte/dotfiles/blob/master/psqlrc

CREATE VIEW lyspgmon.v_unused_indexes AS
  SELECT 
    schemaname AS table_schema, 
    relname AS table_name, 
    indexrelname AS index_name, 
    pg_relation_size(i.indexrelid) AS index_size, 
    pg_size_pretty(pg_relation_size(i.indexrelid)) AS index_size_pretty, 
    idx_scan as index_scans, 
    COALESCE(last_idx_scan,'0001-01-01 12:00:00') AS last_idx_scan
  FROM pg_stat_user_indexes ui
  JOIN pg_index i ON ui.indexrelid = i.indexrelid 
  WHERE NOT indisunique AND idx_scan < 1000 AND pg_relation_size(relid) > 5 * 8192;
