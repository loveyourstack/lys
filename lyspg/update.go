package lyspg

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
)

// getUpdateStmt returns an UPDATE statement using the supplied params
func getUpdateStmt(schemaName, tableName, pkColName string, inputFields []string) string {

	if len(inputFields) == 0 {
		return ""
	}

	var assignments []string
	for k, field := range inputFields {
		assignment := field + " = $" + strconv.Itoa(k+1)
		assignments = append(assignments, assignment)
	}

	return fmt.Sprintf("UPDATE %s.%s SET %s WHERE %s = $%d;",
		schemaName, tableName, strings.Join(assignments, ", "), pkColName, len(inputFields)+1)
}

func getUpdateStmtAndValues[T any, pkT PrimaryKeyType](schemaName, tableName, pkColName string, input T, pkVal pkT, extraDbCols []string, extraInputVals []any) (stmt string, inputVals []any, err error) {

	// if passed, extraDbCols and extraInputVals must be the same length
	if len(extraDbCols) != len(extraInputVals) {
		return "", nil, fmt.Errorf("length of extraDbCols and extraInputVals must be the same")
	}

	// get input values by reflecting input T
	plan, err := lysmeta.AnalyzeValues(input)
	if err != nil {
		return "", nil, fmt.Errorf("lysmeta.AnalyzeValues failed: %w", err)
	}

	dbNames, inputVals, err := plan.DbValues()
	if err != nil {
		return "", nil, fmt.Errorf("plan.DbValues failed: %w", err)
	}

	// add extra columns and values, if any
	dbNames = append(dbNames, extraDbCols...)
	inputVals = append(inputVals, extraInputVals...)

	stmt = getUpdateStmt(schemaName, tableName, pkColName, dbNames)

	// add pkVal as last input value for the WHERE clause
	inputVals = append(inputVals, pkVal)

	return stmt, inputVals, nil
}

// Update changes a single record with the values contained in input
// T must be a struct with "db" tags
func Update[T any, pkT PrimaryKeyType](ctx context.Context, db PoolOrTx, schemaName, tableName, pkColName string, input T, pkVal pkT) error {

	stmt, inputVals, err := getUpdateStmtAndValues(schemaName, tableName, pkColName, input, pkVal, nil, nil)
	if err != nil {
		return fmt.Errorf("getUpdateStmtAndValues failed: %w", err)
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

// UpdateWithExtras works like Update, but allows adding extra columns and values which are not in the input struct, for example to set last_user_update_by from context.
func UpdateWithExtras[T any, pkT PrimaryKeyType](ctx context.Context, db PoolOrTx, schemaName, tableName, pkColName string, input T, pkVal pkT,
	extraDbCols []string, extraInputVals []any) error {

	stmt, inputVals, err := getUpdateStmtAndValues(schemaName, tableName, pkColName, input, pkVal, extraDbCols, extraInputVals)
	if err != nil {
		return fmt.Errorf("getUpdateStmtAndValues failed: %w", err)
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
