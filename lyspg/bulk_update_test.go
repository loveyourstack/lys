package lyspg

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/stretchr/testify/assert"
)

func TestBulkUpdateSuccess(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(t, ctx)
	defer db.Close()

	// delete existing rows, if any
	stmt := "TRUNCATE TABLE core.bulk_update_test;"
	_, err := db.Exec(ctx, stmt)
	if err != nil {
		t.Fatalf("db.Exec (truncate) failed: %v", err)
	}

	// insert 3 records with empty inputs
	emptyInputs := []coretypetestm.Input{}
	for i := 0; i < 3; i++ {
		input := coretypetestm.GetEmptyInput()
		emptyInputs = append(emptyInputs, input)
	}
	_, err = BulkInsert(ctx, db, "core", "bulk_update_test", emptyInputs)
	if err != nil {
		t.Fatalf("BulkInsert failed: %v", err)
	}

	// select ids
	ids, err := SelectArray[int64](ctx, db, "SELECT id FROM core.bulk_update_test ORDER BY id")
	if err != nil {
		t.Fatalf("SelectArray failed: %v", err)
	}
	fmt.Println("ids", ids)

	// prepare 2 filled inputs for the 1st 2 records
	updateInputs := []coretypetestm.Input{}
	for i := 0; i < 2; i++ {
		input, err := coretypetestm.GetFilledInput()
		if err != nil {
			t.Fatalf("coretypetestm.GetFilledInput failed: %v", err)
		}
		updateInputs = append(updateInputs, input)
	}
	pkVals := make([]int64, 2)
	_ = copy(pkVals, ids[:2])

	// add an empty input with an invalid id
	nonExistentId := slices.Max(ids) + 1
	updateInputs = append(updateInputs, coretypetestm.GetEmptyInput())
	pkVals = append(pkVals, nonExistentId)

	// bulk update the first 2 records and should get a partial success msg for the non-existent id
	err = BulkUpdate(ctx, db, "core", "bulk_update_test", "id", updateInputs, pkVals)
	assert.EqualError(t, err, fmt.Sprintf("partial success: invalid pkVals: %d", nonExistentId))

	// test that 1st value was updated
	stmt = "SELECT * FROM core.bulk_update_test WHERE id = $1;"
	rows, _ := db.Query(ctx, stmt, ids[0])
	item, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[coretypetestm.Model])
	if err != nil {
		t.Fatalf("pgx.CollectExactlyOneRow (1st) failed: %v", err)
	}
	coretypetestm.TestFilledInput(t, item.Input)

	// test that 2nd value was updated
	stmt = "SELECT * FROM core.bulk_update_test WHERE id = $1;"
	rows, _ = db.Query(ctx, stmt, ids[1])
	item, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[coretypetestm.Model])
	if err != nil {
		t.Fatalf("pgx.CollectExactlyOneRow (2nd) failed: %v", err)
	}
	coretypetestm.TestFilledInput(t, item.Input)

	// test that 3rd value was not updated
	stmt = "SELECT * FROM core.bulk_update_test WHERE id = $1;"
	rows, _ = db.Query(ctx, stmt, ids[2])
	item, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[coretypetestm.Model])
	if err != nil {
		t.Fatalf("pgx.CollectExactlyOneRow (3rd) failed: %v", err)
	}
	coretypetestm.TestEmptyInput(t, item.Input)
}

func TestBulkUpdateOmitFieldsSuccess(t *testing.T) {
	// TODO
}

func TestBulkUpdateFailure(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(t, ctx)
	defer db.Close()

	// TODO
}
