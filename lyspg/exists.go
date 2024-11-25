package lyspg

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/lyserr"
)

// Exists returns true if 1+ records exist given the supplied criterion (columnName + val)
// pass val = nil to check for NULL
func Exists(ctx context.Context, db PoolOrTx, schemaName, tableName, columnName string, val any) (ret bool, err error) {

	stmt := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s.%s ", schemaName, tableName)

	var rows pgx.Rows
	if val == nil {
		stmt += fmt.Sprintf("WHERE %s IS NULL);", columnName)
		rows, _ = db.Query(ctx, stmt)
	} else {
		stmt += fmt.Sprintf("WHERE %s = $1);", columnName)
		rows, _ = db.Query(ctx, stmt, val)
	}

	ret, err = pgx.CollectExactlyOneRow(rows, pgx.RowTo[bool])
	if err != nil {
		return false, lyserr.Db{Err: fmt.Errorf("pgx.CollectExactlyOneRow failed: %w", err), Stmt: stmt}
	}

	return ret, nil
}

// ExistsConditions returns true if 1+ records exist matching all/any the supplied criteria, depending on match param
// match must be AND or OR
// pass val = nil to check for NULL
func ExistsConditions(ctx context.Context, db PoolOrTx, schemaName, tableName, match string, colValMap map[string]any) (ret bool, err error) {

	match = strings.ToUpper(match)
	if !slices.Contains([]string{"AND", "OR"}, match) {
		return false, fmt.Errorf("match must be 'AND' or 'OR'")
	}

	whereClause := ""
	conds := []string{}
	vals := []any{}
	i := 0

	for col, val := range colValMap {
		var cond string
		if val == nil {
			cond = col + " IS NULL"
		} else {
			i++
			cond = col + " = $" + strconv.Itoa(i)
			vals = append(vals, val)
		}
		conds = append(conds, cond)
	}
	whereClause = strings.Join(conds, " "+match+" ")

	stmt := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s.%s WHERE 1=1 AND (%s));", schemaName, tableName, whereClause)

	rows, _ := db.Query(ctx, stmt, vals...)
	ret, err = pgx.CollectExactlyOneRow(rows, pgx.RowTo[bool])
	if err != nil {
		return false, lyserr.Db{Err: fmt.Errorf("pgx.CollectExactlyOneRow failed: %w", err), Stmt: stmt}
	}

	return ret, nil
}
