package lyspg

import (
	"context"
	"fmt"

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

	// analyze first input for db names
	plan, err := lysmeta.AnalyzeT(inputs[0], false)
	if err != nil {
		return 0, fmt.Errorf("lysmeta.AnalyzeT failed: %w", err)
	}

	// get recs from inputs via reflection
	recs, err := getRecsFromInputs(inputs)
	if err != nil {
		return 0, fmt.Errorf("getRecsFromInputs failed: %w", err)
	}

	// COPY to table using pgx
	rowsAffected, err = db.CopyFrom(ctx, pgx.Identifier{schemaName, tableName}, plan.DbNames(), pgx.CopyFromRows(recs))
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

	// analyze first input for db names
	plan, err := lysmeta.AnalyzeT(inputs[0], false)
	if err != nil {
		return 0, fmt.Errorf("lysmeta.AnalyzeT failed: %w", err)
	}

	recs := getRecsFromInputsWithoutReflection(inputs)

	// COPY to table using pgx
	rowsAffected, err = db.CopyFrom(ctx, pgx.Identifier{"core", "bulk_insert_test"}, plan.DbNames(), pgx.CopyFromRows(recs))
	if err != nil {
		return 0, fmt.Errorf("db.CopyFrom failed: %w", err)
	}

	return rowsAffected, nil
}

func getRecsFromInputs[T any](inputs []T) (recs [][]any, err error) {

	recs = make([][]any, len(inputs))

	for i, input := range inputs {

		// get input values by reflection
		plan, err := lysmeta.AnalyzeT(input, true)
		if err != nil {
			return nil, fmt.Errorf("lysmeta.AnalyzeT failed on input %d: %w", i, err)
		}
		_, inputVals, err := plan.DbValues()
		if err != nil {
			return nil, fmt.Errorf("plan.DbValues failed on input %d: %w", i, err)
		}
		recs[i] = inputVals
	}

	return recs, nil
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
