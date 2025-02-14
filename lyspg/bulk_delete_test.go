package lyspg

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/loveyourstack/lys/internal/stores/core/corebulkdeletetest"
	"github.com/stretchr/testify/assert"
)

func TestBulkDeleteSuccess(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(t, ctx)
	defer db.Close()

	// delete existing rows, if any
	stmt := "TRUNCATE TABLE core.bulk_delete_test;"
	_, err := db.Exec(ctx, stmt)
	if err != nil {
		t.Fatalf("db.Exec (truncate) failed: %v", err)
	}

	// insert 10 records
	inputs := []corebulkdeletetest.Input{}
	for i := 0; i < 10; i++ {
		inputs = append(inputs, corebulkdeletetest.Input{})
	}
	_, err = BulkInsert(ctx, db, "core", "bulk_delete_test", inputs)
	if err != nil {
		t.Fatalf("BulkInsert failed: %v", err)
	}

	// select ids
	ids, err := SelectArray[int64](ctx, db, "SELECT id FROM core.bulk_delete_test")
	if err != nil {
		t.Fatalf("SelectArray failed: %v", err)
	}

	// bulk delete 6 records
	err = BulkDelete(ctx, db, "core", "bulk_delete_test", "id", ids[:6])
	if err != nil {
		t.Fatalf("BulkDelete failed: %v", err)
	}

	// re-select ids
	ids, err = SelectArray[int64](ctx, db, "SELECT id FROM core.bulk_delete_test")
	if err != nil {
		t.Fatalf("SelectArray failed: %v", err)
	}

	// should be 4 left
	assert.EqualValues(t, 4, len(ids), "len(ids) after 1st BulkDelete")

	// append a non-existent id
	nonExistentId := slices.Max(ids) + 1
	ids = append(ids, nonExistentId)

	// bulk delete the 4 remaining valid ids, and should get a partial success msg for the non-existent id
	err = BulkDelete(ctx, db, "core", "bulk_delete_test", "id", ids)
	assert.EqualError(t, err, fmt.Sprintf("partial success: invalid vals: %d", nonExistentId))

	// re-select ids
	ids, err = SelectArray[int64](ctx, db, "SELECT id FROM core.bulk_delete_test")
	if err != nil {
		t.Fatalf("SelectArray failed: %v", err)
	}

	// should be none left
	assert.EqualValues(t, 0, len(ids), "len(ids) after 2nd BulkDelete")
}

func TestBulkDeleteFailure(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(t, ctx)
	defer db.Close()

	// empty input slice
	err := BulkDelete(ctx, db, "core", "bulk_delete_test", "id", []int64{})
	assert.EqualError(t, err, "len(vals) is 0")

	// invalid table
	err = BulkDelete(ctx, db, "core", "bulk_delete_test2", "id", []int64{1})
	assert.EqualError(t, err, `db.SendBatch.Close failed: ERROR: relation "core.bulk_delete_test2" does not exist (SQLSTATE 42P01)`)

	// invalid column
	err = BulkDelete(ctx, db, "core", "bulk_delete_test", "id2", []int64{1})
	assert.EqualError(t, err, `db.SendBatch.Close failed: ERROR: column "id2" does not exist (SQLSTATE 42703)`)
}
