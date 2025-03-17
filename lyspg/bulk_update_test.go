package lyspg

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/internal/stores/core/corearchivetestm"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/loveyourstack/lys/lystype"
	"github.com/stretchr/testify/assert"
)

func TestBulkUpdateSuccess(t *testing.T) {

	schemaName := "core"
	tableName := "bulk_update_test"
	pkColName := "id"

	ctx := context.Background()
	db := mustGetDb(t, ctx)
	defer db.Close()

	// delete existing rows, if any
	stmt := fmt.Sprintf("TRUNCATE TABLE %s.%s;", schemaName, tableName)
	_, err := db.Exec(ctx, stmt)
	if err != nil {
		t.Fatalf("db.Exec (truncate) failed: %v", err)
	}

	// insert 3 records with empty inputs
	emptyInputs := []coretypetestm.Input{}
	for range 3 {
		input := coretypetestm.GetEmptyInput()
		emptyInputs = append(emptyInputs, input)
	}
	_, err = BulkInsert(ctx, db, schemaName, tableName, emptyInputs)
	if err != nil {
		t.Fatalf("BulkInsert failed: %v", err)
	}

	// select ids
	ids, err := SelectArray[int64](ctx, db, fmt.Sprintf("SELECT id FROM %s.%s ORDER BY %s;", schemaName, tableName, pkColName))
	if err != nil {
		t.Fatalf("SelectArray failed: %v", err)
	}
	fmt.Println("ids", ids)

	// prepare 2 filled inputs for the 1st 2 records
	updateInputs := []coretypetestm.Input{}
	for range 2 {
		input, err := coretypetestm.GetFilledInput()
		if err != nil {
			t.Fatalf("coretypetestm.GetFilledInput failed: %v", err)
		}
		updateInputs = append(updateInputs, input)
	}
	pkVals := make([]int64, 2)
	copy(pkVals, ids[:2])

	// add an empty input with an invalid id
	nonExistentId := slices.Max(ids) + 1
	updateInputs = append(updateInputs, coretypetestm.GetEmptyInput())
	pkVals = append(pkVals, nonExistentId)

	// bulk update the first 2 records and should get a partial success msg for the non-existent id
	err = BulkUpdate(ctx, db, schemaName, tableName, pkColName, updateInputs, pkVals)
	assert.EqualError(t, err, fmt.Sprintf("partial success: invalid pkVals: %d", nonExistentId))

	// test that 1st value was updated
	stmt = fmt.Sprintf("SELECT * FROM %s.%s WHERE %s = $1;", schemaName, tableName, pkColName)
	rows, _ := db.Query(ctx, stmt, ids[0])
	item, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[coretypetestm.Model])
	if err != nil {
		t.Fatalf("pgx.CollectExactlyOneRow (1st) failed: %v", err)
	}
	coretypetestm.TestFilledInput(t, item.Input)

	// test that 2nd value was updated
	stmt = fmt.Sprintf("SELECT * FROM %s.%s WHERE %s = $1;", schemaName, tableName, pkColName)
	rows, _ = db.Query(ctx, stmt, ids[1])
	item, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[coretypetestm.Model])
	if err != nil {
		t.Fatalf("pgx.CollectExactlyOneRow (2nd) failed: %v", err)
	}
	coretypetestm.TestFilledInput(t, item.Input)

	// test that 3rd value was not updated
	stmt = fmt.Sprintf("SELECT * FROM %s.%s WHERE %s = $1;", schemaName, tableName, pkColName)
	rows, _ = db.Query(ctx, stmt, ids[2])
	item, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[coretypetestm.Model])
	if err != nil {
		t.Fatalf("pgx.CollectExactlyOneRow (3rd) failed: %v", err)
	}
	coretypetestm.TestEmptyInput(t, item.Input)
}

func TestBulkUpdateOmitFieldsSuccess(t *testing.T) {

	schemaName := "core"
	tableName := "bulk_update_omit_fields_test"
	pkColName := "id"

	ctx := context.Background()
	db := mustGetDb(t, ctx)
	defer db.Close()

	// insert a record
	input := corearchivetestm.Input{
		CInt:  lystype.ToPtr(int64(1)),
		CText: lystype.ToPtr("a"),
	}
	newPk, err := Insert[corearchivetestm.Input, int64](ctx, db, schemaName, tableName, pkColName, input)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	input = corearchivetestm.Input{
		CInt:  lystype.ToPtr(int64(2)),
		CText: lystype.ToPtr("b"),
	}

	// bulk update CInt, but not CText
	err = BulkUpdate(ctx, db, schemaName, tableName, pkColName, []corearchivetestm.Input{input}, []int64{newPk}, UpdateOption{OmitFields: []string{"c_text"}})
	if err != nil {
		t.Fatalf("BulkUpdate failed: %v", err)
	}

	// select record and test
	stmt := fmt.Sprintf("SELECT * FROM %s.%s WHERE %s = $1;", schemaName, tableName, pkColName)
	rows, _ := db.Query(ctx, stmt, newPk)
	item, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[corearchivetestm.Model])
	if err != nil {
		t.Fatalf("pgx.CollectExactlyOneRow failed: %v", err)
	}

	assert.EqualValues(t, 2, *item.CInt, "CInt has changed")
	assert.EqualValues(t, "a", *item.CText, "CText has not changed")
}

func TestBulkUpdateFailure(t *testing.T) {

	schemaName := "core"
	tableName := "bulk_update_test"
	pkColName := "id"

	ctx := context.Background()
	db := mustGetDb(t, ctx)
	defer db.Close()

	type s struct {
		A string
	}

	// empty input slice
	err := BulkUpdate(ctx, db, schemaName, tableName, pkColName, []s{}, []int64{})
	assert.EqualError(t, err, "len(inputs) is 0")

	// len of inputs and pkVals is unequal
	inputs := []s{
		{A: "1"},
	}
	err = BulkUpdate(ctx, db, schemaName, tableName, pkColName, inputs, []int64{})
	assert.EqualError(t, err, "len(inputs) is 1 but len(pkVals) is 0")

	// inputs have no db tags
	err = BulkUpdate(ctx, db, schemaName, tableName, pkColName, inputs, []int64{1})
	assert.EqualError(t, err, "input type does not have db tags")

	type s2 struct {
		A string `db:"a"`
	}
	inputsS2 := []s2{
		{A: "1"},
	}

	// invalid table
	err = BulkUpdate(ctx, db, schemaName, tableName+"2", pkColName, inputsS2, []int64{1})
	assert.EqualError(t, err, `db.SendBatch.Close failed: ERROR: relation "core.bulk_update_test2" does not exist (SQLSTATE 42P01)`)

	// invalid column
	err = BulkUpdate(ctx, db, schemaName, tableName, pkColName+"2", inputsS2, []int64{1})
	assert.EqualError(t, err, `db.SendBatch.Close failed: ERROR: column "id2" does not exist (SQLSTATE 42703)`)
}
