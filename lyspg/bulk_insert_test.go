package lyspg

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/stretchr/testify/assert"
)

// bulkInsertWithoutReflection creates the COPY records for core.bulk_insert_test without using reflection and inserts them
// is used in benchmark (see test file)
func bulkInsertWithoutReflection(ctx context.Context, db PoolOrTx, inputs []coretypetestm.Input) (rowsAffected int64, err error) {

	// check params
	if len(inputs) == 0 {
		return 0, fmt.Errorf("inputs has len 0")
	}

	// analyze first input for db names
	plan, err := lysmeta.Analyze(inputs[0])
	if err != nil {
		return 0, fmt.Errorf("lysmeta.Analyze failed: %w", err)
	}

	recs := getRecsFromInputsWithoutReflection(inputs)

	// COPY to table using pgx
	rowsAffected, err = db.CopyFrom(ctx, pgx.Identifier{"core", "bulk_insert_test"}, plan.DbNames(), pgx.CopyFromRows(recs))
	if err != nil {
		return 0, fmt.Errorf("db.CopyFrom failed: %w", err)
	}

	return rowsAffected, nil
}

// getRecsFromInputsWithoutReflection creates the COPY records for core.bulk_insert_test without using reflection
func getRecsFromInputsWithoutReflection(inputs []coretypetestm.Input) (recs [][]any) {

	recs = make([][]any, len(inputs))

	// directly convert each input to a record without using reflection
	for i, input := range inputs {
		recs[i] = coretypetestm.GetRecord(input)
	}

	return recs
}

func BenchmarkBulkInsert(b *testing.B) {

	schemaName := "core"
	tableName := "bulk_insert_test"

	ctx := context.Background()
	db := mustGetDb(ctx, b)
	defer db.Close()

	b.ResetTimer()

	for range b.N {

		iters := b.N

		b.StopTimer() // don't count creation of inputs

		inputs := make([]coretypetestm.Input, iters)
		for i := range iters {
			input, err := coretypetestm.GetFilledInput()
			if err != nil {
				b.Fatalf("coretypetestm.GetFilledInput failed: %v", err)
			}
			inputs[i] = input
		}

		b.StartTimer()

		_, err := BulkInsert(ctx, db, schemaName, tableName, inputs)
		if err != nil {
			b.Fatalf("BulkInsert failed: %v", err)
		}
	}

	b.StopTimer()

	mustTruncateTable(ctx, b, db, schemaName, tableName)
}

func BenchmarkBulkInsertWithoutReflection(b *testing.B) {

	/*
		This benchmark, when compared with BenchmarkBulkInsert, attempts to measure the use of reflection when creating COPY records for pgx CopyFrom

		I tried comparing getRecsFromInputs with getRecsFromInputsWithoutReflection directly, but the benchmark function didn't return (see below)

		The difference in the results comparing the full functions is not significant though. Method used:

		go test -benchmem -run=^$ -bench ^BenchmarkBulkInsert$ github.com/loveyourstack/lys/lyspg -count=10 | tee stats.txt
		benchstat stats.txt

		go test -benchmem -run=^$ -bench ^BenchmarkBulkInsertWithoutReflection$ github.com/loveyourstack/lys/lyspg -count=10 | tee stats.txt
		benchstat stats.txt

		cleanup: rm stats.txt
	*/

	schemaName := "core"
	tableName := "bulk_insert_test"

	ctx := context.Background()
	db := mustGetDb(ctx, b)
	defer db.Close()

	b.ResetTimer()

	for range b.N {

		iters := b.N

		b.StopTimer() // don't count creation of inputs

		inputs := make([]coretypetestm.Input, iters)
		for i := range iters {
			input, err := coretypetestm.GetFilledInput()
			if err != nil {
				b.Fatalf("coretypetestm.GetFilledInput failed: %v", err)
			}
			inputs[i] = input
		}

		b.StartTimer()

		_, err := bulkInsertWithoutReflection(ctx, db, inputs)
		if err != nil {
			b.Fatalf("bulkInsertWithoutReflection failed: %v", err)
		}
	}

	b.StopTimer()

	mustTruncateTable(ctx, b, db, schemaName, tableName)
}

func BenchmarkGetRecsFromInputs(b *testing.B) {

	// doesn't return within 30s - need to find out why. Tried assigning getRecsFromInputs result to global var, didn't work

	for range b.N {

		iters := b.N

		b.StopTimer() // don't count creation of inputs

		inputs := make([]coretypetestm.Input, iters)
		for i := range iters {
			input, err := coretypetestm.GetFilledInput()
			if err != nil {
				b.Fatalf("coretypetestm.GetFilledInput failed: %v", err)
			}
			inputs[i] = input
		}

		b.StartTimer()

		_, _ = getRecsFromInputs(inputs)
	}
}

func TestBulkInsertSuccess(t *testing.T) {

	schemaName := "core"
	tableName := "bulk_insert_test"
	pkColName := "id"

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	// with empty inputs
	inputs := []coretypetestm.Input{}
	for range 10 {
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
	db := mustGetDb(ctx, t)
	defer db.Close()

	type s struct {
		A string
	}
	inputs := []s{
		{A: "1"},
	}

	// inputs have no db tags
	_, err := BulkInsert(ctx, db, schemaName, tableName, inputs)
	assert.EqualError(t, err, "getRecsFromInputs failed: plan.DbValues failed on input 0: no fields have db tags")

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
