package lyspg

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
)

// getInsertStmt returns an INSERT statement using the supplied params
func getInsertStmt(schemaName, tableName, pkColName string, inputFields []string) string {

	paramPlaceholders := make([]string, len(inputFields))
	for i := range inputFields {
		paramPlaceholders[i] = "$" + strconv.Itoa(i+1)
	}

	return fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s) RETURNING %s;",
		schemaName, tableName, strings.Join(inputFields, ", "), strings.Join(paramPlaceholders, ", "), pkColName)
}

func getInsertStmtAndValues[inputT any](schemaName, tableName, pkColName string, input inputT, extraDbCols []string, extraInputVals []any) (stmt string, inputVals []any, err error) {

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

	stmt = getInsertStmt(schemaName, tableName, pkColName, dbNames)

	return stmt, inputVals, nil
}

// Insert inserts a single record and then returns the new primary key, whose type is pkT
// inputT must be a struct with "db" tags
func Insert[inputT any, pkT PrimaryKeyType](ctx context.Context, db PoolOrTx, schemaName, tableName, pkColName string, input inputT) (newPk pkT, err error) {

	stmt, inputVals, err := getInsertStmtAndValues(schemaName, tableName, pkColName, input, nil, nil)
	if err != nil {
		return newPk, fmt.Errorf("getInsertStmtAndValues failed: %w", err)
	}

	if err = db.QueryRow(ctx, stmt, inputVals...).Scan(&newPk); err != nil {
		return newPk, lyserr.Db{Err: fmt.Errorf(ErrDescInsertScanFailed+": %w", err), Stmt: stmt}
	}

	return newPk, nil
}

// InsertSelect inserts a single record and then returns it
// inputT must be a struct with "db" tags
func InsertSelect[inputT any, itemT any](ctx context.Context, db PoolOrTx, schemaName, tableName, viewName, pkColName string, input inputT) (newItem itemT, err error) {

	stmt, inputVals, err := getInsertStmtAndValues(schemaName, tableName, pkColName, input, nil, nil)
	if err != nil {
		return newItem, fmt.Errorf("getInsertStmtAndValues failed: %w", err)
	}

	var newPk any
	if err = db.QueryRow(ctx, stmt, inputVals...).Scan(&newPk); err != nil {
		return newItem, lyserr.Db{Err: fmt.Errorf(ErrDescInsertScanFailed+": %w", err), Stmt: stmt}
	}

	return SelectUnique[itemT](ctx, db, schemaName, viewName, pkColName, newPk)
}

// InsertWithExtras works like Insert, but allows adding extra columns and values which are not in the input struct, for example to set created_by from context.
func InsertWithExtras[inputT any, pkT PrimaryKeyType](ctx context.Context, db PoolOrTx, schemaName, tableName, pkColName string, input inputT,
	extraDbCols []string, extraInputVals []any) (newPk pkT, err error) {

	stmt, inputVals, err := getInsertStmtAndValues(schemaName, tableName, pkColName, input, extraDbCols, extraInputVals)
	if err != nil {
		return newPk, fmt.Errorf("getInsertStmtAndValues failed: %w", err)
	}

	if err = db.QueryRow(ctx, stmt, inputVals...).Scan(&newPk); err != nil {
		return newPk, lyserr.Db{Err: fmt.Errorf(ErrDescInsertScanFailed+": %w", err), Stmt: stmt}
	}

	return newPk, nil
}
