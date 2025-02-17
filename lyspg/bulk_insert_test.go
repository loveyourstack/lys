package lyspg

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/stretchr/testify/assert"
)

func TestBulkInsertSuccess(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(t, ctx)
	defer db.Close()

	// with empty inputs
	inputs := []coretypetestm.Input{}
	for i := 0; i < 10; i++ {
		input := coretypetestm.GetEmptyInput()
		inputs = append(inputs, input)
	}

	rowsAffected, err := BulkInsert(ctx, db, "core", "bulk_insert_test", inputs)
	if err != nil {
		t.Fatalf("BulkInsert (empty) failed: %v", err)
	}
	assert.EqualValues(t, 10, rowsAffected, "type test - empty")

	// test last inserted value
	stmt := "SELECT * FROM core.bulk_insert_test ORDER BY id DESC LIMIT 1"
	rows, _ := db.Query(ctx, stmt)
	item, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[coretypetestm.Model])
	if err != nil {
		t.Fatalf("pgx.CollectExactlyOneRow failed: %v", err)
	}
	coretypetestm.TestEmptyInput(t, item.Input)

	// with filled inputs
	inputs = []coretypetestm.Input{}
	for i := 0; i < 10; i++ {
		input, err := coretypetestm.GetFilledInput()
		if err != nil {
			t.Fatalf("coretypetestm.GetFilledInput failed: %v", err)
		}
		inputs = append(inputs, input)
	}

	rowsAffected, err = BulkInsert(ctx, db, "core", "bulk_insert_test", inputs)
	if err != nil {
		t.Fatalf("BulkInsert (filled) failed: %v", err)
	}
	assert.EqualValues(t, 10, rowsAffected, "type test - filled")

	// test last inserted value
	stmt = "SELECT * FROM core.bulk_insert_test ORDER BY id DESC LIMIT 1"
	rows, _ = db.Query(ctx, stmt)
	item, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[coretypetestm.Model])
	if err != nil {
		t.Fatalf("pgx.CollectExactlyOneRow failed: %v", err)
	}
	coretypetestm.TestFilledInput(t, item.Input)
}

func TestBulkInsertFailure(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(t, ctx)
	defer db.Close()

	// inputs have no db tags
	type s struct {
		A string
	}
	inputs := []s{
		{A: "1"},
	}
	_, err := BulkInsert(ctx, db, "core", "bulk_insert_test", inputs)
	assert.EqualError(t, err, "input type does not have db tags")

	// empty input slice
	_, err = BulkInsert(ctx, db, "core", "bulk_insert_test", []s{})
	assert.EqualError(t, err, "inputs has len 0")
}
