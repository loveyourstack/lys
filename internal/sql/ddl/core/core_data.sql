
INSERT INTO core.archive_test (c_int, c_text) VALUES
  (1, NULL),
  (NULL, 'a')
;

INSERT INTO core.param_test (c_bool, c_booln, c_int, c_intn, c_double, c_doublen, c_date, c_daten, c_time, c_timen, c_datetime, c_datetimen, c_enum, c_enumn, c_text, c_textn) VALUES
  (false, NULL, 1, 2, 1.1, NULL, '2001-01-01', NULL, '12:01', NULL, '2001-01-01 12:01:00+01', NULL, 'Monday', NULL, 'a', ''),
  (true, true, 2, 2, 2.1, 2.1, '2002-01-01', '2002-01-01', '12:02', '12:02', '2002-01-01 12:01:00+01', '2002-01-01 12:01:00+01', 'Tuesday', 'Tuesday', 'b', 'abc')
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


INSERT INTO core.volume_test (c_rnd, c_int) 
  -- 1 mil rows (100,000 * 10)
  SELECT (random()*100)::int, b FROM generate_series(1,100000) a, generate_series(1,10) b;
ANALYZE core.volume_test;