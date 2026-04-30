package lyspg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/lyserr"
)

// ValueMap returns of map of k = keyCol, v = valCol from the supplied table. KeyCol must be unique.
// Conds may be supplied to filter the rows that are included in the map. If conds is nil or empty, all rows will be included.
func ValueMap[keyT comparable, valT any](ctx context.Context, db PoolOrTx, schemaName, tableName, keyCol, valCol string, conds []Condition) (m map[keyT]valT, err error) {

	type s struct {
		K keyT
		V valT
	}

	whereClause, _ := GetWhereClause(0, conds, nil)
	paramVals := GetSelectParamValues(nil, conds, nil, false, 0, 0)

	stmt := fmt.Sprintf("SELECT %s, %s FROM %s.%s WHERE 1=1%s;", keyCol, valCol, schemaName, tableName, whereClause)
	rows, _ := db.Query(ctx, stmt, paramVals...)
	items, err := pgx.CollectRows(rows, pgx.RowToStructByPos[s])
	if err != nil {
		return nil, lyserr.Db{Err: fmt.Errorf("pgx.CollectRows failed: %w", err), Stmt: stmt}
	}

	m = make(map[keyT]valT, len(items))
	for _, item := range items {
		m[item.K] = item.V
	}

	if len(items) != len(m) {
		return nil, fmt.Errorf("%s.%s: key '%s' is not unique. len(items) is %v, but len(m) is %v", schemaName, tableName, keyCol, len(items), len(m))
	}

	return m, nil
}
