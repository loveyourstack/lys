package lyspg

import (
	"context"
	"fmt"
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/loveyourstack/lys/lysmeta"
)

// BulkInsert inserts multiple records using the postgres COPY protocol
// T must be a struct with "db" tags
func BulkInsert[T any](ctx context.Context, db PoolOrTx, schemaName, tableName string, inputs []T) (rowsAffected int64, err error) {

	// check params
	if len(inputs) == 0 {
		return 0, fmt.Errorf("inputs has len 0")
	}

	// get db tags of first input
	inputReflVals := reflect.ValueOf(inputs[0])
	meta, err := lysmeta.AnalyzeStructs(inputReflVals)
	if err != nil {
		return 0, fmt.Errorf("lysmeta.AnalyzeStructs failed: %w", err)
	}
	if len(meta.DbTags) == 0 {
		return 0, fmt.Errorf("input type does not have db tags")
	}

	// get recs from inputs via reflection
	recs := getRecsFromInputs(inputs)

	// COPY to table using pgx
	rowsAffected, err = db.CopyFrom(ctx, pgx.Identifier{schemaName, tableName}, meta.DbTags, pgx.CopyFromRows(recs))
	if err != nil {
		return 0, fmt.Errorf("db.CopyFrom failed: %w", err)
	}

	return rowsAffected, nil
}

// bulkInsertWithoutReflection creates the COPY records for core.bulk_insert_test without using reflection and inserts them
// is used in benchmark (see test file)
func bulkInsertWithoutReflection(ctx context.Context, db PoolOrTx, inputs []coretypetestm.Input) (rowsAffected int64, err error) {

	// check params
	if len(inputs) == 0 {
		return 0, fmt.Errorf("inputs has len 0")
	}

	// get db tags of first input
	inputReflVals := reflect.ValueOf(inputs[0])
	meta, err := lysmeta.AnalyzeStructs(inputReflVals)
	if err != nil {
		return 0, fmt.Errorf("lysmeta.AnalyzeStructs failed: %w", err)
	}
	if len(meta.DbTags) == 0 {
		return 0, fmt.Errorf("input type does not have db tags")
	}

	recs := getRecsFromInputsWithoutReflection(inputs)

	// COPY to table using pgx
	rowsAffected, err = db.CopyFrom(ctx, pgx.Identifier{"core", "bulk_insert_test"}, meta.DbTags, pgx.CopyFromRows(recs))
	if err != nil {
		return 0, fmt.Errorf("db.CopyFrom failed: %w", err)
	}

	return rowsAffected, nil
}

func getRecsFromInputs[T any](inputs []T) (recs [][]any) {

	recs = make([][]any, len(inputs))

	for i, input := range inputs {

		// get input values by reflection
		inputReflVals := reflect.ValueOf(input)
		recs[i] = getInputValsFromStruct(inputReflVals, nil)
	}

	return recs
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
