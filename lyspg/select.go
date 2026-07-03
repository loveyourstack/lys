package lyspg

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/lyserr"
)

// Select returns multiple rows from the db according to the params supplied
func Select[T any](ctx context.Context, db PoolOrTx, schemaName, tableName, viewName, defaultOrderBy string, allFields []string,
	params SelectParams) (items []T, unpagedCount TotalCount, err error) {

	// use allFields if the fields param was not sent
	var fields []string
	if params.Fields == nil {
		fields = allFields
	} else {
		fields = params.Fields
	}

	// escape pg reserved words in fields
	EscapeReserved(fields)

	// build select stmt with placeholders for conditions
	selectCols := strings.Join(fields, ",")
	whereClause, numPlaceholders := GetWhereClause(len(params.SetFuncParamValues), params.Conditions, params.OrConditionSets)
	sourceName := GetSourceName(viewName, len(params.SetFuncParamValues))
	stmt := GetSelectStem(selectCols, schemaName, sourceName, whereClause)

	// if unpagedCount is requested
	if params.GetUnpagedCount && tableName != "" {

		// get a fast recordcount
		unpagedCount, err = fastRowCount(ctx, db, schemaName, tableName, params.SetFuncParamValues, params.Conditions, params.OrConditionSets, stmt)
		if err != nil {
			return nil, TotalCount{}, fmt.Errorf("fastRowCount failed: %w", err)
		}
	}

	stmt += GetOrderBy(params.Sorts, defaultOrderBy)
	stmt += GetLimitOffsetClause(numPlaceholders)

	// get params for stmt placeholders
	paramValues := GetSelectParamValues(params.SetFuncParamValues, params.Conditions, params.OrConditionSets, true, GetLimit(params.Limit), params.Offset)

	// using RowToStructByNameLax below because the fields param might restrict the number of columns selected
	// causing a mismatch between # of columns returned and the # of fields in the dest struct

	rows, _ := db.Query(ctx, stmt, paramValues...)
	items, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])
	if err != nil {
		return nil, TotalCount{}, lyserr.Db{Err: fmt.Errorf("pgx.CollectRows failed: %w", err), Stmt: stmt}
	}

	// success
	return items, unpagedCount, nil
}

// SelectSlice is a wrapper for selecting into a non-struct type T (db.Query / pgx.CollectRows with RowTo).
// T must be a primitive type such as int64 or string.
func SelectSlice[T any](ctx context.Context, db PoolOrTx, selectStmt string, params ...any) (ar []T, err error) {

	rows, _ := db.Query(ctx, selectStmt, params...)
	ar, err = pgx.CollectRows(rows, pgx.RowTo[T])
	if err != nil {
		return nil, lyserr.Db{Err: fmt.Errorf("pgx.CollectRows failed: %w", err), Stmt: selectStmt}
	}

	return ar, nil
}

// SelectBySlice returns multiple rows from the db depending on the slice supplied.
// inputT must be a primitive type such as int64 or string.
func SelectBySlice[inputT, itemT any](ctx context.Context, db PoolOrTx, schema, view, column string, ar []inputT) (items []itemT, err error) {

	stmt := fmt.Sprintf("SELECT * FROM %s.%s WHERE %s = ANY($1);", schema, view, column)

	rows, _ := db.Query(ctx, stmt, ar)
	items, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[itemT])
	if err != nil {
		return nil, lyserr.Db{Err: fmt.Errorf("pgx.CollectRows failed: %w", err), Stmt: stmt}
	}

	return items, nil
}

// SelectT is a wrapper for selecting into a struct type T (db.Query / pgx.CollectRows with RowToStructByNameLax)
func SelectT[T any](ctx context.Context, db PoolOrTx, selectStmt string, params ...any) (items []T, err error) {

	rows, _ := db.Query(ctx, selectStmt, params...)
	items, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])
	if err != nil {
		return nil, lyserr.Db{Err: fmt.Errorf("pgx.CollectRows failed: %w", err), Stmt: selectStmt}
	}

	return items, nil
}

// SelectUniqueRowFields returns the requested fields of a single row using the value of a unique column such as id.
// fields is a slice of column names to select. If you want all columns, use "*".
func SelectUniqueRowFields[T any](ctx context.Context, db PoolOrTx, fields []string, schema, view, uniqueCol string, uniqueVal any) (item T, err error) {

	stmt := fmt.Sprintf(`SELECT %s FROM %s.%s WHERE %s = $1;`, strings.Join(fields, ", "), schema, view, uniqueCol)

	rows, _ := db.Query(ctx, stmt, uniqueVal)
	item, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[T])
	if err != nil {
		return item, lyserr.Db{Err: fmt.Errorf("pgx.CollectExactlyOneRow failed: %w", err), Stmt: stmt}
	}

	// success
	return item, nil
}

// SelectUnique returns a single row using the value of a unique column such as id.
func SelectUnique[T any](ctx context.Context, db PoolOrTx, schema, view, uniqueCol string, uniqueVal any) (item T, err error) {
	return SelectUniqueRowFields[T](ctx, db, []string{"*"}, schema, view, uniqueCol, uniqueVal)
}
