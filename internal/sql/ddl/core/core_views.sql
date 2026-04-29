
CREATE OR REPLACE VIEW core.v_tag_test AS
  SELECT
    id,
    c_editable,
    c_hidden,
    c_obscured
  FROM core.tag_test;