package lyspg

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
)

// BulkUpdate changes multiple records in the same table in a single pg batch. The records are identified by pkVals with the values contained in inputs
// T must be a struct with "db" tags
// partial success possible: if some pkVals are not found, an error will be returned containing the failed pks, but the other rows will be updated
func BulkUpdate[T any, pkT PrimaryKeyType](ctx context.Context, db PoolOrTx, schemaName, tableName, pkColName string, inputs []T, pkVals []pkT,
	options ...UpdateOption) error {

	if len(inputs) == 0 {
		return fmt.Errorf("len(inputs) is %v", len(inputs))
	}
	if len(inputs) != len(pkVals) {
		return fmt.Errorf("len(inputs) is %v but len(pkVals) is %v", len(inputs), len(pkVals))
	}

	// get columns to update by reflecting the first input T
	inputReflVals := reflect.ValueOf(inputs[0])
	meta, err := lysmeta.AnalyzeStructs(inputReflVals)
	if err != nil {
		return fmt.Errorf("lysmeta.AnalyzeStructs failed: %w", err)
	}
	if len(meta.DbTags) == 0 {
		return fmt.Errorf("input type does not have db tags")
	}

	// get fields to omit from the update, if any
	omitFields := getOmitFields(options...)

	// get updateFields (dbTags with omitted fields removed)
	updateFields := getUpdateFields(meta.DbTags, omitFields)

	stmt := getUpdateStmt(schemaName, tableName, pkColName, updateFields)
	batch := &pgx.Batch{}
	invalidPkVals := []pkT{}

	// for each record to be updated
	for i := range inputs {

		// get input values by reflecting input
		inputReflVals := reflect.ValueOf(inputs[i])
		inputVals := getInputValsFromStruct(inputReflVals, omitFields)

		// add pk as final input val
		inputVals = append(inputVals, pkVals[i])

		// queue the query
		batch.Queue(stmt, inputVals...).Exec(func(ct pgconn.CommandTag) error {
			if ct.RowsAffected() == 0 {
				invalidPkVals = append(invalidPkVals, pkVals[i])
			}
			return nil
		})
	}

	// send all queries to db
	// any SQL syntax errors will fail here and no rows will be updated
	err = db.SendBatch(ctx, batch).Close()
	if err != nil {
		return lyserr.Db{Err: fmt.Errorf("db.SendBatch.Close failed: %w", err)}
	}

	// if some pkVals were invalid, return them
	if len(invalidPkVals) > 0 {
		rets := make([]string, len(invalidPkVals))
		for i, pkVal := range invalidPkVals {
			rets[i] = fmt.Sprintf("%v", pkVal)
		}
		return lyserr.Db{Err: fmt.Errorf("partial success: invalid pkVals: %s", strings.Join(rets, ", "))}
	}

	return nil
}
