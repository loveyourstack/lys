package lyspg

import (
	"context"
	"fmt"
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
)

// UpdatePartial updates only the supplied columns of the record
// assignmentsMap is a map of k = json key, v = new value
func UpdatePartial[pkT PrimaryKeyType](ctx context.Context, db PoolOrTx, schemaName, tableName, pkColName string, jsonKeyDbNameMap map[string]string, assignmentsMap map[string]any, pkVal pkT) error {

	fmt.Printf("%+v", assignmentsMap)

	// get db column names and input values from assignmentsMap
	dbNames := make([]string, len(assignmentsMap))
	inputVals := make([]any, len(assignmentsMap))
	i := 0
	for k, v := range assignmentsMap {

		dbName, ok := jsonKeyDbNameMap[k]
		if !ok {
			return lyserr.User{Message: fmt.Sprintf("invalid field: %s", k)}
		}
		dbNames[i] = dbName
		inputVals[i] = lysmeta.GetInputValue(v, reflect.TypeOf(v))
		i++
	}

	stmt := getUpdateStmt(schemaName, tableName, pkColName, dbNames)

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
func UpdatePartialWithLastUserBy[pkT PrimaryKeyType](ctx context.Context, db PoolOrTx, schemaName, tableName, pkColName string, jsonKeyDbNameMap map[string]string,
	assignmentsMap map[string]any, pkVal pkT, lastUserUpdateBy string) error {

	assignmentsMap["last_user_update_by"] = lastUserUpdateBy
	jsonKeyDbNameMap["last_user_update_by"] = "last_user_update_by"
	return UpdatePartial(ctx, db, schemaName, tableName, pkColName, jsonKeyDbNameMap, assignmentsMap, pkVal)
}
