package lyspg

import (
	"context"
	"slices"
	"testing"
)

func TestGetChildForeignKeys(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	fks, err := GetChildForeignKeys(ctx, db, "core", "info_schema_parent_test")
	if err != nil {
		t.Fatalf("GetChildForeignKeys failed: %v", err)
	}

	if len(fks) != 1 {
		t.Fatalf("unexpected FK count: got %d, want 1", len(fks))
	}
	if fks[0].ChildTable != "info_schema_child_test" {
		t.Errorf("unexpected child table: got %s, want info_schema_child_test", fks[0].ChildTable)
	}
	if fks[0].ChildColumn != "c_parent_fk" {
		t.Errorf("unexpected child column: got %s, want c_parent_fk", fks[0].ChildColumn)
	}
	if fks[0].ParentSchema != "core" {
		t.Errorf("unexpected parent schema: got %s, want core", fks[0].ParentSchema)
	}
}

func TestGetForeignKeys(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	fks, err := GetForeignKeys(ctx, db, "core", "info_schema_child_test")
	if err != nil {
		t.Fatalf("GetForeignKeys failed: %v", err)
	}

	if len(fks) != 1 {
		t.Fatalf("unexpected FK count: got %d, want 1", len(fks))
	}
	if fks[0].ParentTable != "info_schema_parent_test" {
		t.Errorf("unexpected parent table: got %s, want info_schema_parent_test", fks[0].ParentTable)
	}
	if fks[0].ChildColumn != "c_parent_fk" {
		t.Errorf("unexpected child column: got %s, want c_parent_fk", fks[0].ChildColumn)
	}
	if fks[0].ParentColumn != "id" {
		t.Errorf("unexpected parent column: got %s, want id", fks[0].ParentColumn)
	}
}

func TestGetStatsRowCount(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	rowCount, err := GetStatsRowCount(ctx, db, "core", "info_schema_parent_test")
	if err != nil {
		t.Fatalf("GetStatsRowCount failed: %v", err)
	}

	// reltuples is -1 for unanalyzed tables and >= 0 after ANALYZE; both are valid
	if rowCount < -1 {
		t.Errorf("unexpected row count: got %d", rowCount)
	}
}

func TestGetTableColumns(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	// ORDER BY in GetTableColumns: is_identity DESC, column_name → id, c_text, c_parent_fk
	cols, err := GetTableColumns(ctx, db, "core", "info_schema_child_test")
	if err != nil {
		t.Fatalf("GetTableColumns failed: %v", err)
	}

	if len(cols) != 3 {
		t.Fatalf("unexpected column count: got %d, want 3", len(cols))
	}

	// id must be the first column (identity)
	idCol := cols[0]
	if idCol.Name != "id" {
		t.Errorf("unexpected first column: got %s, want id", idCol.Name)
	}
	if !idCol.IsIdentity {
		t.Errorf("expected id to be identity column")
	}
	if idCol.IsNullable {
		t.Errorf("expected id to be not nullable")
	}
	if idCol.IsTracking {
		t.Errorf("expected id to not be a tracking column")
	}

	// c_parent_fk must be NOT NULL and not identity
	fkColIdx := slices.IndexFunc(cols, func(c Column) bool { return c.Name == "c_parent_fk" })
	if fkColIdx == -1 {
		t.Fatalf("c_parent_fk column not found")
	}
	if cols[fkColIdx].IsIdentity {
		t.Errorf("expected c_parent_fk to not be identity")
	}
	if cols[fkColIdx].IsNullable {
		t.Errorf("expected c_parent_fk to be not nullable")
	}
}

func TestGetTableColumnNames(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	schemaName := "core"
	tableName := "info_schema_child_test"

	colNames, err := GetTableColumnNames(ctx, db, schemaName, tableName)
	if err != nil {
		t.Fatalf("GetTableColumnNames failed: %v", err)
	}

	expectedColNames := []string{"id", "c_parent_fk", "c_text"}
	if !slices.Equal(colNames, expectedColNames) {
		t.Fatalf("unexpected col names: got %v, want %v", colNames, expectedColNames)
	}
}

func TestGetTableComment(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	comment, err := GetTableComment(ctx, db, "core", "info_schema_parent_test")
	if err != nil {
		t.Fatalf("GetTableComment failed: %v", err)
	}

	expected := "shortname: isp"
	if comment != expected {
		t.Errorf("unexpected comment: got %q, want %q", comment, expected)
	}
}

func TestGetTableShortName(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	shortName, err := GetTableShortName(ctx, db, "core", "info_schema_parent_test")
	if err != nil {
		t.Fatalf("GetTableShortName failed: %v", err)
	}

	expected := "isp"
	if shortName != expected {
		t.Errorf("unexpected short name: got %q, want %q", shortName, expected)
	}
}

func TestGetTableShortName_noComment(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	// archive_test has no COMMENT ON TABLE → should return empty string, not error
	shortName, err := GetTableShortName(ctx, db, "core", "archive_test")
	if err != nil {
		t.Fatalf("GetTableShortName failed: %v", err)
	}

	if shortName != "" {
		t.Errorf("expected empty short name for table without comment, got %q", shortName)
	}
}
