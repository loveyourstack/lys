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

// Insert inserts a single record and then returns the new primary key, whose type is pkT
// inputT must be a struct with "db" tags
func Insert[inputT any, pkT PrimaryKeyType](ctx context.Context, db PoolOrTx, schemaName, tableName, pkColName string, input inputT) (newPk pkT, err error) {

	// get input values by reflecting input T
	plan, err := lysmeta.AnalyzeValues(input)
	if err != nil {
		return newPk, fmt.Errorf("lysmeta.AnalyzeValues failed: %w", err)
	}

	dbNames, inputVals, err := plan.DbValues()
	if err != nil {
		return newPk, fmt.Errorf("plan.DbValues failed: %w", err)
	}

	stmt := getInsertStmt(schemaName, tableName, pkColName, dbNames)

	if err = db.QueryRow(ctx, stmt, inputVals...).Scan(&newPk); err != nil {
		return newPk, lyserr.Db{Err: fmt.Errorf(ErrDescInsertScanFailed+": %w", err), Stmt: stmt}
	}

	return newPk, nil
}

// InsertSelect inserts a single record and then returns it
// inputT must be a struct with "db" tags
func InsertSelect[inputT any, itemT any](ctx context.Context, db PoolOrTx, schemaName, tableName, viewName, pkColName string, input inputT) (newItem itemT, err error) {

	// get input values by reflecting input T
	plan, err := lysmeta.AnalyzeValues(input)
	if err != nil {
		return newItem, fmt.Errorf("lysmeta.AnalyzeValues failed: %w", err)
	}

	dbNames, inputVals, err := plan.DbValues()
	if err != nil {
		return newItem, fmt.Errorf("plan.DbValues failed: %w", err)
	}

	stmt := getInsertStmt(schemaName, tableName, pkColName, dbNames)

	var newPk any
	if err = db.QueryRow(ctx, stmt, inputVals...).Scan(&newPk); err != nil {
		return newItem, lyserr.Db{Err: fmt.Errorf(ErrDescInsertScanFailed+": %w", err), Stmt: stmt}
	}

	return SelectUnique[itemT](ctx, db, schemaName, viewName, pkColName, newPk)
}

// InsertWithCreatedBy works like Insert, but adds a created_by field to the input struct and sets it to the supplied createdBy value
func InsertWithCreatedBy[inputT any, pkT PrimaryKeyType](ctx context.Context, db PoolOrTx, schemaName, tableName, pkColName string, input inputT, createdBy string) (newPk pkT, err error) {

	// get input values by reflecting input T
	plan, err := lysmeta.AnalyzeValues(input)
	if err != nil {
		return newPk, fmt.Errorf("lysmeta.AnalyzeValues failed: %w", err)
	}

	dbNames, inputVals, err := plan.DbValues()
	if err != nil {
		return newPk, fmt.Errorf("plan.DbValues failed: %w", err)
	}

	// add created_by
	dbNames = append(dbNames, "created_by")
	inputVals = append(inputVals, createdBy)

	stmt := getInsertStmt(schemaName, tableName, pkColName, dbNames)

	if err = db.QueryRow(ctx, stmt, inputVals...).Scan(&newPk); err != nil {
		return newPk, lyserr.Db{Err: fmt.Errorf(ErrDescInsertScanFailed+": %w", err), Stmt: stmt}
	}

	return newPk, nil
}
