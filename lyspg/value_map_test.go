package lyspg

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/loveyourstack/lys/lyserr"
)

func TestValueMap_intKey(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	m, err := ValueMap[int64, string](ctx, db, "core", "value_map_test", "id", "c_text", nil)
	if err != nil {
		t.Fatalf("ValueMap failed: %v", err)
	}

	if len(m) != 3 {
		t.Fatalf("unexpected map len: got %d, want 3", len(m))
	}
	if m[1] != "alpha" {
		t.Errorf("m[1]: got %q, want \"alpha\"", m[1])
	}
	if m[2] != "beta" {
		t.Errorf("m[2]: got %q, want \"beta\"", m[2])
	}
	if m[3] != "gamma" {
		t.Errorf("m[3]: got %q, want \"gamma\"", m[3])
	}
}

func TestValueMap_stringKey(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	m, err := ValueMap[string, int64](ctx, db, "core", "value_map_test", "c_text", "id", nil)
	if err != nil {
		t.Fatalf("ValueMap failed: %v", err)
	}

	if len(m) != 3 {
		t.Fatalf("unexpected map len: got %d, want 3", len(m))
	}
	if m["alpha"] != 1 {
		t.Errorf("m[\"alpha\"]: got %d, want 1", m["alpha"])
	}
}

func TestValueMap_withConds(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	conds := []Condition{{Field: "c_int", Operator: OpGreaterThan, Value: "1"}}
	m, err := ValueMap[int64, string](ctx, db, "core", "value_map_test", "id", "c_text", conds)
	if err != nil {
		t.Fatalf("ValueMap failed: %v", err)
	}

	if len(m) != 2 {
		t.Fatalf("unexpected map len: got %d, want 2", len(m))
	}
	if _, ok := m[2]; !ok {
		t.Errorf("expected key 2 in map")
	}
	if _, ok := m[3]; !ok {
		t.Errorf("expected key 3 in map")
	}
}

func TestValueMap_emptyResult(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	conds := []Condition{{Field: "c_int", Operator: OpGreaterThan, Value: "100"}}
	m, err := ValueMap[int64, string](ctx, db, "core", "value_map_test", "id", "c_text", conds)
	if err != nil {
		t.Fatalf("ValueMap failed: %v", err)
	}

	if len(m) != 0 {
		t.Fatalf("expected empty map, got len %d", len(m))
	}
}

func TestValueMap_nonUniqueKey(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	// core.exists_test has c_int values 1,1,2,3 — duplicate key 1 triggers the uniqueness check
	// use id (bigint, never null) as the value column to avoid NULL scan errors on c_text
	_, err := ValueMap[int64, int64](ctx, db, "core", "exists_test", "c_int", "id", nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "not unique") {
		t.Errorf("expected error to contain \"not unique\", got: %v", err)
	}

	// must NOT be a lyserr.Db — it is a generic fmt.Errorf
	var dbErr lyserr.Db
	if errors.As(err, &dbErr) {
		t.Errorf("expected generic error, got lyserr.Db")
	}
}

func TestValueMap_invalidColumn(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	_, err := ValueMap[int64, string](ctx, db, "core", "value_map_test", "no_such_col", "c_text", nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var dbErr lyserr.Db
	if !errors.As(err, &dbErr) {
		t.Fatalf("expected lyserr.Db, got %T: %v", err, err)
	}
}
