package lyspg

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/stretchr/testify/assert"
)

func TestInsertSelect_returnedItemMatchesDbRecord(t *testing.T) {

	insertTestSchema := "core"
	insertTestTable := "type_test"
	insertTestPk := "id"

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	input := coretypetestm.GetEmptyInput()
	input.CText = "round-trip check"

	item, err := InsertSelect[coretypetestm.Input, coretypetestm.Model](ctx, db, insertTestSchema, insertTestTable, insertTestTable, insertTestPk, input)
	if err != nil {
		t.Fatalf("InsertSelect failed: %v", err)
	}

	// fetch the same record directly and compare
	stmt := fmt.Sprintf("SELECT * FROM %s.%s WHERE %s = $1;", insertTestSchema, insertTestTable, insertTestPk)
	rows, _ := db.Query(ctx, stmt, item.Id)
	dbItem, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[coretypetestm.Model])
	if err != nil {
		t.Fatalf("pgx.CollectExactlyOneRow failed: %v", err)
	}

	assert.EqualValues(t, dbItem, item, "returned item matches db record")
}

func TestInsertSelect_invalidInput(t *testing.T) {

	insertTestSchema := "core"
	insertTestTable := "type_test"
	insertTestPk := "id"

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	// CEnum is required; passing an invalid value should cause a DB error
	input := coretypetestm.GetEmptyInput()
	input.CEnum = "NotAWeekday"

	_, err := InsertSelect[coretypetestm.Input, coretypetestm.Model](ctx, db, insertTestSchema, insertTestTable, insertTestTable, insertTestPk, input)
	assert.Error(t, err, "expected error for invalid enum value")
}
