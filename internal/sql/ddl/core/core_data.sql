
INSERT INTO core.archive_test (c_int, c_text) VALUES
  (1, NULL),
  (NULL, 'a')
;

INSERT INTO core.archive_test_uuid (id, c_int, c_text) VALUES
  ('b872908f-68ca-4c87-96dd-98a9b97470f0', 1, NULL),
  ('4bfc505c-c316-4385-af4f-63160e7326d9', NULL, 'a')
;

INSERT INTO core.exists_test (c_date, c_int, c_text) VALUES
  ('2001-01-01', 1, 'a'),
  (NULL, 1, 'a'),
  ('2002-01-01', 2, 'b'),
  ('2003-01-01', 3, NULL)
;

INSERT INTO core.param_test (c_bool, c_booln, c_int, c_intn, c_double, c_doublen, c_date, c_daten, c_time, c_timen, c_datetime, c_datetimen, c_enum, c_enumn, c_text, c_textn) VALUES
  (false, NULL, 1, 2, 1.1, NULL, '2001-01-01', NULL, '12:01', NULL, '2001-01-01 12:01:00+01', NULL, 'Monday', NULL, 'a', ''),
  (true, true, 2, 2, 2.1, 2.1, '2002-01-01', '2002-01-01', '12:02', '12:02', '2002-01-01 12:01:00+01', '2002-01-01 12:01:00+01', 'Tuesday', 'Tuesday', 'b', 'abc')
;

INSERT INTO core.info_schema_parent_test (c_text) VALUES ('a');
INSERT INTO core.info_schema_child_test (c_parent_fk, c_text) VALUES 
  (1, 'a'),
  (1, 'b'),
  (1, 'c');


INSERT INTO core.tag_test (c_editable, c_hidden, c_obscured) VALUES
  ('e1', 'h1', 'o1'), -- for get tests
  ('e2', 'h2', 'o2'), -- for put tests
  ('e3', 'h3', 'o3') -- for patch tests
;

INSERT INTO core.tracking_test (c_editable, created_by, last_user_update_by) VALUES
  ('e1', 'insert', ''), -- for put tests
  ('e2', 'insert', '') -- for patch tests
;

INSERT INTO core.type_test (c_bool, c_boola, c_int, c_inta, c_double, c_doublea, c_numeric, c_numerica, c_date, c_datea, c_time, c_timea, c_datetime, c_datetimea, c_enum, c_enuma, c_text, c_texta) VALUES (
	true, '{true,false}',
  1, '{1,2}',
  1.1, '{1.1,2.1}',
  1.11, '{1.11,2.11}',
  '2001-02-03', '{2001-02-03,2002-03-04}',
  '12:01', '{12:01,12:02}',
  '2001-02-03 12:01:00+01', '{2001-02-03 12:01:00+01,2002-03-04 12:02:00+01}',
  'Monday', '{Monday,Tuesday}',
  'a b', '{a b,b c}'
);

INSERT INTO core.uuid_test (id, c_int, c_text) VALUES
  ('b872908f-68ca-4c87-96dd-98a9b97470f0', 1, 'a'), -- for get tests
  ('4bfc505c-c316-4385-af4f-63160e7326d9', 2, 'b'), -- for put tests
  ('0707b293-b39c-4ae6-8ab9-a92b592e9568', 3, 'c') -- for patch tests
;

INSERT INTO core.value_map_test (c_text, c_int) VALUES
  ('alpha', 1),
  ('beta', 2),
  ('gamma', 3)
;


INSERT INTO core.volume_test (c_rnd, c_int) 
  -- 1 mil rows (100,000 * 10)
  SELECT (random()*100)::int, b FROM generate_series(1,100000) a, generate_series(1,10) b;


ANALYZE;