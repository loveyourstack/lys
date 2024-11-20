package lyspg

import (
	"context"
	"fmt"
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
)

// BulkUpdate changes multiple records in the same table identified by pkVals with the values contained in inputs
// T must be a struct with "db" tags
func BulkUpdate[T any, pkT PrimaryKeyType](ctx context.Context, db PoolOrTx, schemaName, tableName, pkColName string, inputs []T, pkVals []pkT,
	options ...UpdateOption) (rowsAffected int64, err error) {

	if len(inputs) == 0 {
		return 0, fmt.Errorf("len(inputs) is %v", len(inputs))
	}
	if len(inputs) != len(pkVals) {
		return 0, fmt.Errorf("len(inputs) is %v but len(pkVals) is %v", len(inputs), len(pkVals))
	}

	// get columns to update by reflecting the first input T
	inputReflVals := reflect.ValueOf(inputs[0])
	meta, err := lysmeta.AnalyzeStructs(inputReflVals)
	if err != nil {
		return 0, fmt.Errorf("lysmeta.AnalyzeStructs failed: %w", err)
	}

	// get fields to omit from the update, if any
	omitFields := getOmitFields(options...)

	// get updateFields (dbTags with omitted fields removed)
	updateFields := getUpdateFields(meta.DbTags, omitFields)

	stmt := getUpdateStmt(schemaName, tableName, pkColName, updateFields)
	batch := &pgx.Batch{}

	// for each record to be updated
	for i := range inputs {

		// get input values by reflecting input
		inputReflVals := reflect.ValueOf(inputs[i])
		inputVals := getInputValsFromStruct(inputReflVals, omitFields)

		// add pk as final input val
		inputVals = append(inputVals, pkVals[i])

		// queue the query
		batch.Queue(stmt, inputVals...)
	}

	// send all queries to db
	batchRes := db.SendBatch(ctx, batch)
	defer batchRes.Close()

	// exec all queries
	cmdTag, err := batchRes.Exec()
	if err != nil {
		return 0, lyserr.Db{Err: fmt.Errorf("batchRes.Exec (update %v recs) failed: %w", batch.Len(), err)}
	}

	if cmdTag.RowsAffected() == 0 {
		return 0, pgx.ErrNoRows
	}

	return cmdTag.RowsAffected(), nil
}
