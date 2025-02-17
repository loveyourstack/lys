
CREATE TABLE core.archive_test
(
  id bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  id_uu uuid NOT NULL DEFAULT gen_random_uuid(),
  c_int int,
  c_text text,
  entry_at tracking_at
);
--- change columns with ---
CREATE TABLE core.archive_test_archived
(
  id bigint NOT NULL PRIMARY KEY,
  id_uu uuid,
  c_int int,
  c_text text,
  entry_at tracking_at,
  archived_at tracking_at,
  archived_by_cascade bool NOT NULL
);


CREATE TABLE core.exists_test (LIKE core.archive_test INCLUDING ALL);


CREATE TABLE core.bulk_update_omit_fields_test (LIKE core.archive_test INCLUDING ALL);


CREATE TABLE core.bulk_delete_test
(
  id bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  c_text text NOT NULL,
  entry_at tracking_at
);


CREATE TABLE core.param_test
(
  id bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  id_uu uuid NOT NULL DEFAULT gen_random_uuid(),
  c_bool bool NOT NULL,
  c_booln bool,
  c_int int NOT NULL,
  c_intn int,
  c_double double precision NOT NULL,
  c_doublen double precision,
  c_date date NOT NULL,
  c_daten date,
  c_time time NOT NULL,
  c_timen time,
  c_datetime timestamptz NOT NULL,
  c_datetimen timestamptz,
  c_enum core.weekday NOT NULL DEFAULT 'None',
  c_enumn core.weekday,
  c_text text NOT NULL,
  c_textn text
);


CREATE TABLE core.type_test
(	
  id bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  id_uu UUID DEFAULT gen_random_uuid(),
  c_bool bool NOT NULL DEFAULT false,
  c_booln bool,
  c_boola bool[] NOT NULL DEFAULT '{}',
  c_int int NOT NULL,
  c_intn int,
  c_inta int[] NOT NULL DEFAULT '{}',
  c_double double precision NOT NULL,
  c_doublen double precision,
  c_doublea double precision[] NOT NULL DEFAULT '{}',
  c_numeric numeric(12,2) NOT NULL,
  c_numericn numeric(12,2),
  c_numerica numeric(12,2)[] NOT NULL DEFAULT '{}',
  c_date date NOT NULL,
  c_daten date,
  c_datea date[] NOT NULL DEFAULT '{}',
  c_time time NOT NULL,
  c_timen time,
  c_timea time[] NOT NULL DEFAULT '{}',
  c_datetime timestamptz NOT NULL,
  c_datetimen timestamptz,
  c_datetimea timestamptz[] NOT NULL DEFAULT '{}',
  c_enum core.weekday NOT NULL DEFAULT 'None',
  c_enumn core.weekday,
  c_enuma core.weekday[] NOT NULL DEFAULT '{}',
  c_text text NOT NULL,
  c_textn text,
  c_texta text[] NOT NULL DEFAULT '{}'
);
COMMENT ON TABLE core.type_test IS 'shortname: tt';


CREATE TABLE core.bulk_insert_test (LIKE core.type_test INCLUDING ALL);


CREATE TABLE core.bulk_update_test (LIKE core.type_test INCLUDING ALL);


CREATE TABLE core.volume_test
(
  id bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  c_rnd int NOT NULL,
  c_int int NOT NULL
);