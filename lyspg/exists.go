package lyspg

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// Exists returns true if 1+ records exist given the supplied criterion (columnName + val)
func Exists(ctx context.Context, db PoolOrTx, schemaName, tableName, columnName string, val any) (ret bool, stmt string, err error) {

	stmt = fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s.%s WHERE %s = $1);", schemaName, tableName, columnName)

	rows, _ := db.Query(ctx, stmt, val)
	ret, err = pgx.CollectExactlyOneRow(rows, pgx.RowTo[bool])
	if err != nil {
		return false, stmt, fmt.Errorf("pgx.CollectExactlyOneRow failed: %w", err)
	}

	return ret, "", nil
}

// ExistsConditions returns true if 1+ records exist matching all/any the supplied criteria, depending on match param
// match must be AND or OR
func ExistsConditions(ctx context.Context, db PoolOrTx, schemaName, tableName, match string, colValMap map[string]any) (ret bool, stmt string, err error) {

	if !slices.Contains([]string{"AND", "OR"}, match) {
		return false, "", fmt.Errorf("match must be 'AND' or 'OR'")
	}

	whereClause := ""
	conds := []string{}
	vals := []any{}
	i := 0

	for col, val := range colValMap {
		i++
		cond := col + " = $" + strconv.Itoa(i)
		conds = append(conds, cond)
		vals = append(vals, val)
	}
	whereClause = strings.Join(conds, " "+match+" ")

	stmt = fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s.%s WHERE 1=1 AND (%s));", schemaName, tableName, whereClause)

	rows, _ := db.Query(ctx, stmt, vals...)
	ret, err = pgx.CollectExactlyOneRow(rows, pgx.RowTo[bool])
	if err != nil {
		return false, stmt, fmt.Errorf("pgx.CollectExactlyOneRow failed: %w", err)
	}

	return ret, "", nil
}
