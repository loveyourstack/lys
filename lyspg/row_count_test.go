package lyspg

import (
	"context"
	"errors"
	"testing"

	"github.com/loveyourstack/lys/lyserr"
)

func TestGetRowCount(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	rowCount, err := GetRowCount(ctx, db, "core", "exists_test")
	if err != nil {
		t.Fatalf("GetRowCount failed: %v", err)
	}

	if rowCount != 4 {
		t.Fatalf("unexpected row count: got %d, want 4", rowCount)
	}
}

func TestGetRowCountPlaceholderQry(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	qry := "SELECT * FROM core.exists_test WHERE c_text = $1"
	rowCount, err := GetRowCountPlaceholderQry(ctx, db, qry, []any{"a"})
	if err != nil {
		t.Fatalf("GetRowCountPlaceholderQry failed: %v", err)
	}

	if rowCount != 2 {
		t.Fatalf("unexpected row count: got %d, want 2", rowCount)
	}
}

func TestGetRowCountPlaceholderQry_invalidPlaceholder(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	qry := "SELECT * FROM core.exists_test WHERE c_text = $2"
	_, err := GetRowCountPlaceholderQry(ctx, db, qry, []any{"a"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var dbErr lyserr.Db
	if !errors.As(err, &dbErr) {
		t.Fatalf("expected lyserr.Db, got %T", err)
	}
}

func TestGetRowCountExplain(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	qry := "SELECT * FROM core.exists_test WHERE c_text = $1"
	rowCount, err := GetRowCountExplain(ctx, db, qry, []any{"a"})
	if err != nil {
		t.Fatalf("GetRowCountExplain failed: %v", err)
	}

	if rowCount < 0 {
		t.Fatalf("unexpected estimated row count: got %d", rowCount)
	}
}

func TestGetRowCountExplain_invalidPlaceholder(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	qry := "SELECT * FROM core.exists_test WHERE c_text = $2"
	_, err := GetRowCountExplain(ctx, db, qry, []any{"a"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var dbErr lyserr.Db
	if !errors.As(err, &dbErr) {
		t.Fatalf("expected lyserr.Db, got %T", err)
	}
}

func TestFastRowCount_smallTableNoConds(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	totalCount, err := fastRowCount(ctx, db, "core", "exists_test", nil, nil, nil, "SELECT * FROM core.exists_test")
	if err != nil {
		t.Fatalf("fastRowCount failed: %v", err)
	}

	if totalCount.IsEstimated {
		t.Fatalf("expected exact row count, got estimated")
	}
	if totalCount.Value != 4 {
		t.Fatalf("unexpected row count: got %d, want 4", totalCount.Value)
	}
}

func TestFastRowCount_smallTableWithConds(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	conds := []Condition{{
		Field:    "c_text",
		Operator: OpEquals,
		Value:    "a",
	}}

	totalCount, err := fastRowCount(ctx, db, "core", "exists_test", nil, conds, nil, "SELECT * FROM core.exists_test WHERE 1=1 AND c_text = $1")
	if err != nil {
		t.Fatalf("fastRowCount failed: %v", err)
	}

	if totalCount.IsEstimated {
		t.Fatalf("expected exact row count, got estimated")
	}
	if totalCount.Value != 2 {
		t.Fatalf("unexpected row count: got %d, want 2", totalCount.Value)
	}
}

func TestFastRowCount_largeTableNoConds(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	statsRowCount, err := GetStatsRowCount(ctx, db, "core", "volume_test")
	if err != nil {
		t.Fatalf("GetStatsRowCount failed: %v", err)
	}
	if statsRowCount <= largeTableThreshold {
		t.Fatalf("expected volume_test stats to exceed threshold, got %d", statsRowCount)
	}

	totalCount, err := fastRowCount(ctx, db, "core", "volume_test", nil, nil, nil, "SELECT * FROM core.volume_test")
	if err != nil {
		t.Fatalf("fastRowCount failed: %v", err)
	}

	if !totalCount.IsEstimated {
		t.Fatalf("expected estimated row count, got exact")
	}
	if totalCount.Value != statsRowCount {
		t.Fatalf("unexpected row count: got %d, want stats %d", totalCount.Value, statsRowCount)
	}
}

func TestFastRowCount_largeTableWithConds(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	statsRowCount, err := GetStatsRowCount(ctx, db, "core", "volume_test")
	if err != nil {
		t.Fatalf("GetStatsRowCount failed: %v", err)
	}
	if statsRowCount <= largeTableThreshold {
		t.Fatalf("expected volume_test stats to exceed threshold, got %d", statsRowCount)
	}

	conds := []Condition{{
		Field:    "c_int",
		Operator: OpEquals,
		Value:    "1",
	}}

	totalCount, err := fastRowCount(ctx, db, "core", "volume_test", nil, conds, nil, "SELECT * FROM core.volume_test WHERE 1=1 AND c_int = $1")
	if err != nil {
		t.Fatalf("fastRowCount failed: %v", err)
	}

	if !totalCount.IsEstimated {
		t.Fatalf("expected estimated row count, got exact")
	}
	if totalCount.Value < 0 {
		t.Fatalf("unexpected estimated row count: got %d", totalCount.Value)
	}
}

func TestSelectCount_noConds(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	totalCount, err := SelectCount(ctx, db, "core", "exists_test", "exists_test", SelectParams{})
	if err != nil {
		t.Fatalf("SelectCount failed: %v", err)
	}

	if totalCount.IsEstimated {
		t.Fatalf("expected exact count for small table, got estimated")
	}
	if totalCount.Value != 4 {
		t.Fatalf("unexpected count: got %d, want 4", totalCount.Value)
	}
}

func TestSelectCount_withCond(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	params := SelectParams{
		Conditions: []Condition{{
			Field:    "c_text",
			Operator: OpEquals,
			Value:    "a",
		}},
	}

	totalCount, err := SelectCount(ctx, db, "core", "exists_test", "exists_test", params)
	if err != nil {
		t.Fatalf("SelectCount failed: %v", err)
	}

	if totalCount.IsEstimated {
		t.Fatalf("expected exact count for small table, got estimated")
	}
	if totalCount.Value != 2 {
		t.Fatalf("unexpected count: got %d, want 2", totalCount.Value)
	}
}

func TestSelectCount_withCondMatchesNone(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	params := SelectParams{
		Conditions: []Condition{{
			Field:    "c_text",
			Operator: OpEquals,
			Value:    "zzz",
		}},
	}

	totalCount, err := SelectCount(ctx, db, "core", "exists_test", "exists_test", params)
	if err != nil {
		t.Fatalf("SelectCount failed: %v", err)
	}

	if totalCount.Value != 0 {
		t.Fatalf("unexpected count: got %d, want 0", totalCount.Value)
	}
}

func TestSelectCount_largeTableNoConds(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	totalCount, err := SelectCount(ctx, db, "core", "volume_test", "volume_test", SelectParams{})
	if err != nil {
		t.Fatalf("SelectCount failed: %v", err)
	}

	if !totalCount.IsEstimated {
		t.Fatalf("expected estimated count for large table, got exact")
	}
	if totalCount.Value <= 100_000 {
		t.Fatalf("unexpected count: got %d", totalCount.Value)
	}
}

func TestSelectCount_largeTableWithCond(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	params := SelectParams{
		Conditions: []Condition{{
			Field:    "c_int",
			Operator: OpEquals,
			Value:    "1",
		}},
	}

	totalCount, err := SelectCount(ctx, db, "core", "volume_test", "volume_test", params)
	if err != nil {
		t.Fatalf("SelectCount failed: %v", err)
	}

	if !totalCount.IsEstimated {
		t.Fatalf("expected estimated count for large table with conds, got exact")
	}
	if totalCount.Value <= 50_000 {
		t.Fatalf("unexpected estimated count: got %d", totalCount.Value)
	}
}
