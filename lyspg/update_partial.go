package lyspg

import (
	"context"
	"fmt"
	"reflect"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
)

// UpdatePartial updates only the supplied columns of the record
// assignmentsMap is a map of k = column name, v = new value
func UpdatePartial[pkT PrimaryKeyType](ctx context.Context, db PoolOrTx, schemaName, tableName, pkColName string, allowedFields []string, assignmentsMap map[string]any, pkVal pkT) error {

	// get keys (column names) and input values from assignmentsMap
	keys := make([]string, len(assignmentsMap))
	inputVals := make([]any, len(assignmentsMap))
	i := 0
	for k, v := range assignmentsMap {
		//fmt.Printf("%s: %v\n", k, v)
		keys[i] = k
		inputVals[i] = lysmeta.GetInputValue(v, reflect.TypeOf(v))
		i++
	}

	// ensure that each map key is among the allowed fields
	for _, k := range keys {
		if !slices.Contains(allowedFields, k) {
			return lyserr.User{Message: fmt.Sprintf("invalid field: %s", k)}
		}
	}

	stmt := getUpdateStmt(schemaName, tableName, pkColName, keys)

	inputVals = append(inputVals, pkVal)

	cmdTag, err := db.Exec(ctx, stmt, inputVals...)
	if err != nil {
		return lyserr.Db{Err: fmt.Errorf(ErrDescUpdateExecFailed+": %w", err), Stmt: stmt}
	}

	if cmdTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	// success
	return nil
}

// UpdatePartialWithLastUserBy is a wrapper for UpdatePartial that adds a last_user_update_by field to the assignmentsMap and sets it to the supplied lastUserUpdateBy value
func UpdatePartialWithLastUserBy[pkT PrimaryKeyType](ctx context.Context, db PoolOrTx, schemaName, tableName, pkColName string, allowedFields []string,
	assignmentsMap map[string]any, pkVal pkT, lastUserUpdateBy string) error {

	assignmentsMap["last_user_update_by"] = lastUserUpdateBy
	return UpdatePartial(ctx, db, schemaName, tableName, pkColName, append(allowedFields, "last_user_update_by"), assignmentsMap, pkVal)
}
