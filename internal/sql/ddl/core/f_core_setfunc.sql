
CREATE OR REPLACE FUNCTION core.setfunc (
	_p_text text,
  _p_int int,
  _p_inta int[]
)
RETURNS TABLE (
  text_val text,
  int_val int
) AS
$BODY$

WITH unnested AS (
  SELECT unnest(_p_inta) AS res
)

SELECT _p_text, res + _p_int
FROM unnested;

$BODY$
  LANGUAGE sql VOLATILE
  COST 100;
