
DROP VIEW IF EXISTS lyspgmon.v_settings;

CREATE VIEW lyspgmon.v_settings AS
  SELECT 
    name,
    COALESCE(setting,'') AS setting,
    COALESCE(boot_val,'') AS boot_val,
    COALESCE(unit,'') AS unit,
    context,
    short_desc,
    COALESCE(extra_desc,'') AS extra_desc,
    CASE WHEN COALESCE(setting,'') != COALESCE(boot_val,'') THEN true ELSE false END AS changed
  FROM pg_settings;
