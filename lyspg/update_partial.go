package lyspg

import (
	"context"
	"fmt"
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
)

func getUpdatePartialStmtAndValues[pkT PrimaryKeyType](schemaName, tableName, pkColName string, jsonKeyDbNameMap map[string]string, assignmentsMap map[string]any,
	pkVal pkT, extraDbCols []string, extraInputVals []any) (stmt string, inputVals []any, err error) {

	// if passed, extraDbCols and extraInputVals must be the same length
	if len(extraDbCols) != len(extraInputVals) {
		return "", nil, fmt.Errorf("length of extraDbCols and extraInputVals must be the same")
	}

	// get db column names and input values from assignmentsMap
	dbNames := make([]string, len(assignmentsMap))
	inputVals = make([]any, len(assignmentsMap))
	i := 0
	for k, v := range assignmentsMap {

		dbName, ok := jsonKeyDbNameMap[k]
		if !ok {
			return "", nil, lyserr.User{Message: fmt.Sprintf("invalid field: %s", k)}
		}
		dbNames[i] = dbName
		inputVals[i] = lysmeta.GetInputValue(v, reflect.TypeOf(v))
		i++
	}

	// add extra columns and values, if any
	dbNames = append(dbNames, extraDbCols...)
	inputVals = append(inputVals, extraInputVals...)

	stmt = getUpdateStmt(schemaName, tableName, pkColName, dbNames)

	// add pkVal as last input value for the WHERE clause
	inputVals = append(inputVals, pkVal)

	return stmt, inputVals, nil
}

// UpdatePartial updates only the supplied columns of the record
// assignmentsMap is a map of k = json key, v = new value
func UpdatePartial[pkT PrimaryKeyType](ctx context.Context, db PoolOrTx, schemaName, tableName, pkColName string, jsonKeyDbNameMap map[string]string, assignmentsMap map[string]any, pkVal pkT) error {

	stmt, inputVals, err := getUpdatePartialStmtAndValues(schemaName, tableName, pkColName, jsonKeyDbNameMap, assignmentsMap, pkVal, nil, nil)
	if err != nil {
		return fmt.Errorf("getUpdatePartialStmtAndValues failed: %w", err)
	}

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

// UpdatePartialWithExtras is a wrapper for UpdatePartial that adds extra columns and values to the update statement, for example to set last_user_update_by from context
func UpdatePartialWithExtras[pkT PrimaryKeyType](ctx context.Context, db PoolOrTx, schemaName, tableName, pkColName string, jsonKeyDbNameMap map[string]string,
	assignmentsMap map[string]any, pkVal pkT, extraDbCols []string, extraInputVals []any) error {

	stmt, inputVals, err := getUpdatePartialStmtAndValues(schemaName, tableName, pkColName, jsonKeyDbNameMap, assignmentsMap, pkVal, extraDbCols, extraInputVals)
	if err != nil {
		return fmt.Errorf("getUpdatePartialStmtAndValues failed: %w", err)
	}

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
