package lyspg

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/stretchr/testify/assert"
)

func TestBulkInsertSuccess(t *testing.T) {

	schemaName := "core"
	tableName := "bulk_insert_test"
	pkColName := "id"

	ctx := context.Background()
	db := mustGetDb(t, ctx)
	defer db.Close()

	// with empty inputs
	inputs := []coretypetestm.Input{}
	for i := 0; i < 10; i++ {
		input := coretypetestm.GetEmptyInput()
		inputs = append(inputs, input)
	}

	rowsAffected, err := BulkInsert(ctx, db, schemaName, tableName, inputs)
	if err != nil {
		t.Fatalf("BulkInsert (empty) failed: %v", err)
	}
	assert.EqualValues(t, 10, rowsAffected, "type test - empty")

	// test last inserted value
	stmt := fmt.Sprintf("SELECT * FROM %s.%s ORDER BY %s DESC LIMIT 1", schemaName, tableName, pkColName)
	rows, _ := db.Query(ctx, stmt)
	item, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[coretypetestm.Model])
	if err != nil {
		t.Fatalf("pgx.CollectExactlyOneRow failed: %v", err)
	}
	coretypetestm.TestEmptyInput(t, item.Input)

	// with filled inputs
	inputs = []coretypetestm.Input{}
	for range 10 {
		input, err := coretypetestm.GetFilledInput()
		if err != nil {
			t.Fatalf("coretypetestm.GetFilledInput failed: %v", err)
		}
		inputs = append(inputs, input)
	}

	rowsAffected, err = BulkInsert(ctx, db, schemaName, tableName, inputs)
	if err != nil {
		t.Fatalf("BulkInsert (filled) failed: %v", err)
	}
	assert.EqualValues(t, 10, rowsAffected, "type test - filled")

	// test last inserted value
	stmt = fmt.Sprintf("SELECT * FROM %s.%s ORDER BY %s DESC LIMIT 1", schemaName, tableName, pkColName)
	rows, _ = db.Query(ctx, stmt)
	item, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[coretypetestm.Model])
	if err != nil {
		t.Fatalf("pgx.CollectExactlyOneRow failed: %v", err)
	}
	coretypetestm.TestFilledInput(t, item.Input)
}

func TestBulkInsertFailure(t *testing.T) {

	schemaName := "core"
	tableName := "bulk_insert_test"

	ctx := context.Background()
	db := mustGetDb(t, ctx)
	defer db.Close()

	type s struct {
		A string
	}
	inputs := []s{
		{A: "1"},
	}

	// inputs have no db tags
	_, err := BulkInsert(ctx, db, schemaName, tableName, inputs)
	assert.EqualError(t, err, "input type does not have db tags")

	// empty input slice
	_, err = BulkInsert(ctx, db, schemaName, tableName, []s{})
	assert.EqualError(t, err, "inputs has len 0")

	type s2 struct {
		A string `db:"a"`
	}
	inputsS2 := []s2{
		{A: "1"},
	}

	// invalid table
	_, err = BulkInsert(ctx, db, schemaName, tableName+"2", inputsS2)
	assert.EqualError(t, err, `db.CopyFrom failed: statement description failed: ERROR: relation "core.bulk_insert_test2" does not exist (SQLSTATE 42P01)`)

	// invalid column
	_, err = BulkInsert(ctx, db, schemaName, tableName, inputsS2)
	assert.EqualError(t, err, `db.CopyFrom failed: statement description failed: ERROR: column "a" does not exist (SQLSTATE 42703)`)
}
