package lyspg

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// Select returns multiple rows from the db according to the params supplied
func Select[T any](ctx context.Context, db PoolOrTx, schemaName, tableName, viewName, defaultOrderBy string, allFields []string,
	params SelectParams) (items []T, unpagedCount TotalCount, stmt string, err error) {

	// use allFields if the fields param was not sent
	var fields []string
	if params.Fields == nil {
		fields = allFields
	} else {
		fields = params.Fields
	}

	// build select stmt with placeholders for conditions
	selectCols := strings.Join(fields, ",")
	whereClause, numPlaceholders := getWhereClause(params.Conditions)
	stmt = getSelectStem(selectCols, schemaName, viewName, whereClause)

	// if unpagedCount is requested
	if params.GetUnpagedCount && tableName != "" {

		// get a fast recordcount
		var errStmt string
		unpagedCount, errStmt, err = fastRowCount(ctx, db, schemaName, tableName, params.Conditions, stmt)
		if err != nil {
			return nil, TotalCount{}, errStmt, fmt.Errorf("fastRowCount failed: %w", err)
		}
	}

	stmt += getOrderBy(params.Sorts, defaultOrderBy)
	stmt += getLimitOffsetClause(numPlaceholders)

	// get params for stmt placeholders
	paramValues := getSelectParamValues(params.Conditions, true, getLimit(params.Limit), params.Offset)

	// using RowToStructByNameLax below because the fields param might restrict the number of columns selected
	// causing a mismatch between # of columns returned and the # of fields in the dest struct

	//fmt.Println(stmt)
	rows, _ := db.Query(ctx, stmt, paramValues...)
	items, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])
	if err != nil {
		return nil, TotalCount{}, stmt, fmt.Errorf("pgx.CollectRows failed: %w", err)
	}

	// success
	return items, unpagedCount, stmt, nil
}

// SelectByArray returns multiple rows from the db depending on the array supplied
// inputT must be a primitive type such as int64 or string
func SelectByArray[inputT, itemT any](ctx context.Context, db PoolOrTx, schema, view, column string, ar []inputT) (items []itemT, stmt string, err error) {

	stmt = fmt.Sprintf("SELECT * FROM %s.%s WHERE %s = ANY($1);", schema, view, column)

	rows, _ := db.Query(ctx, stmt, ar)
	items, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[itemT])
	if err != nil {
		return nil, stmt, fmt.Errorf("pgx.CollectRows failed: %w", err)
	}

	return items, "", nil
}

// SelectUnique returns a single row using the value of a unique column such as id
func SelectUnique[T any](ctx context.Context, db PoolOrTx, schema, view, column string, fields, allFields []string, uniqueVal any) (item T, stmt string, err error) {

	// use allFields if the fields param was not sent
	if fields == nil {
		fields = allFields
	}

	stmt = fmt.Sprintf(`SELECT %s FROM %s.%s WHERE %s = $1;`, strings.Join(fields, ","), schema, view, column)

	rows, _ := db.Query(ctx, stmt, uniqueVal)
	item, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[T])
	if err != nil {
		return item, stmt, fmt.Errorf("pgx.CollectExactlyOneRow failed: %w", err)
	}

	// success
	return item, stmt, nil
}
