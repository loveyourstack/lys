package lyspgdb

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestEnhanceMigrationFile(t *testing.T) {
	migrationAssets := fstest.MapFS{
		"nested/rewrite.sql": {
			Data: []byte("-- migration header\n-- + public.f_example;\nSELECT 1;\n-- + public.v_example"),
		},
	}
	ddlAssets := fstest.MapFS{
		"public/funcs/f_example.sql": {Data: []byte("CREATE FUNCTION example();\n")},
		"public/views/v_example.sql": {Data: []byte("CREATE VIEW example_view AS SELECT 1;\n")},
	}

	got, err := EnhanceMigrationFile("rewrite.sql", ddlAssets, migrationAssets)
	if err != nil {
		t.Fatalf("EnhanceMigrationFile() error = %v", err)
	}

	want := "BEGIN;\n-- migration header\nCREATE FUNCTION example();\n\nSELECT 1;\nCREATE VIEW example_view AS SELECT 1;\n\nROLLBACK;\n--COMMIT;"
	if got != want {
		t.Errorf("EnhanceMigrationFile() = %q, want %q", got, want)
	}
}

func TestEnhanceMigrationFileErrors(t *testing.T) {
	tests := []struct {
		name            string
		fileName        string
		migrationAssets fstest.MapFS
		ddlAssets       fstest.MapFS
		wantErr         string
	}{
		{
			name:     "missing migration",
			fileName: "missing.sql",
			wantErr:  "no match found for missing.sql",
		},
		{
			name:     "missing object",
			fileName: "missing-object.sql",
			migrationAssets: fstest.MapFS{
				"missing-object.sql": {Data: []byte("-- + public.f_missing")},
			},
			wantErr: "fs.ReadFile failed for public/funcs/f_missing.sql",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EnhanceMigrationFile(tt.fileName, tt.ddlAssets, tt.migrationAssets)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("EnhanceMigrationFile() error = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestGetMigrationFile(t *testing.T) {
	migrationAssets := fstest.MapFS{
		"nested/rewrite.sql": {Data: []byte("-- migration contents")},
	}

	got, err := getMigrationFile("rewrite.sql", migrationAssets)
	if err != nil {
		t.Fatalf("getMigrationFile() error = %v", err)
	}
	if want := "-- migration contents"; got != want {
		t.Errorf("getMigrationFile() = %q, want %q", got, want)
	}
}

func TestGetMigrationFileErrors(t *testing.T) {
	tests := []struct {
		name            string
		fileName        string
		migrationAssets fstest.MapFS
		wantErr         string
	}{
		{
			name:     "no match",
			fileName: "absent.sql",
			wantErr:  "no match found for absent.sql",
		},
		{
			name:     "multiple matches",
			fileName: "duplicate.sql",
			migrationAssets: fstest.MapFS{
				"first/duplicate.sql":  {Data: []byte("first")},
				"second/duplicate.sql": {Data: []byte("second")},
			},
			wantErr: "expected exactly one match for duplicate.sql, found 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getMigrationFile(tt.fileName, tt.migrationAssets)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("getMigrationFile() error = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestGetMigrationObjectReplacement(t *testing.T) {
	ddlAssets := fstest.MapFS{
		"public/funcs/f_example.sql":  {Data: []byte("CREATE FUNCTION example();")},
		"public/funcs/p_example.sql":  {Data: []byte("CREATE PROCEDURE example();")},
		"public/funcs/tf_example.sql": {Data: []byte("CREATE FUNCTION table_example();")},
		"public/views/mv_example.sql": {Data: []byte("CREATE MATERIALIZED VIEW example_materialized AS SELECT 1;")},
		"public/views/v_example.sql":  {Data: []byte("CREATE VIEW example_view AS SELECT 1;")},
	}
	tests := []struct {
		name string
		obj  string
		want string
	}{
		{name: "function", obj: "public.f_example", want: "CREATE FUNCTION example();"},
		{name: "procedure", obj: "public.p_example", want: "CREATE PROCEDURE example();"},
		{name: "table function", obj: "public.tf_example", want: "CREATE FUNCTION table_example();"},
		{name: "materialized view", obj: "public.mv_example", want: "CREATE MATERIALIZED VIEW example_materialized AS SELECT 1;"},
		{name: "view", obj: "public.v_example", want: "CREATE VIEW example_view AS SELECT 1;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getMigrationObjectReplacement(tt.obj, ddlAssets)
			if err != nil {
				t.Fatalf("getMigrationObjectReplacement() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("getMigrationObjectReplacement() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetMigrationObjectReplacementErrors(t *testing.T) {
	tests := []struct {
		name    string
		obj     string
		wantErr string
	}{
		{name: "missing schema separator", obj: "f_example", wantErr: "invalid object name: f_example"},
		{name: "unsupported prefix", obj: "public.t_example", wantErr: "unsupported object name: t_example"},
		{name: "schema path separator", obj: "bad/schema.f_example", wantErr: "invalid object name: bad/schema.f_example"},
		{name: "schema backslash", obj: `bad\schema.f_example`, wantErr: `invalid object name: bad\schema.f_example`},
		{name: "object path separator", obj: "public.f_../example", wantErr: "invalid object name: public.f_../example"},
		{name: "object backslash", obj: `public.f_\example`, wantErr: `invalid object name: public.f_\example`},
		{name: "missing object file", obj: "public.f_missing", wantErr: "fs.ReadFile failed for public/funcs/f_missing.sql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getMigrationObjectReplacement(tt.obj, fstest.MapFS{})
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("getMigrationObjectReplacement() error = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}
