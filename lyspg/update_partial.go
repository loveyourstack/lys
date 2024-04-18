package lyspg

import (
	"context"
	"fmt"
	"reflect"
	"slices"

	"github.com/jackc/pgx/v5"
)

// UpdatePartial updates only the supplied columns of the record
// assignmentsMap is a map of k = column name, v = new value
// pkVal is type "any" so that both int and string PKs can be used
func UpdatePartial(ctx context.Context, db PoolOrTx, schemaName, tableName, pkColName string, allowedFields []string, assignmentsMap map[string]any,
	pkVal any) (stmt string, err error) {

	// get keys (column names) and input values from assignmentsMap
	var keys []string
	var inputVals []any
	for k, v := range assignmentsMap {
		keys = append(keys, k)
		inputVals = append(inputVals, getInputValue(v, reflect.TypeOf(v)))
	}

	// ensure that each map key is among the allowed fields
	for _, k := range keys {
		if !slices.Contains(allowedFields, k) {
			return "", fmt.Errorf("invalid field: %s", k)
		}
	}

	stmt = getUpdateStmt(schemaName, tableName, pkColName, keys)

	inputVals = append(inputVals, pkVal)

	cmdTag, err := db.Exec(ctx, stmt, inputVals...)
	if err != nil {
		return stmt, fmt.Errorf(ErrDescUpdateExecFailed+": %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return stmt, pgx.ErrNoRows
	}

	// success
	return "", nil
}
