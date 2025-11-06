package lyspg

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/loveyourstack/lys/lyserr"
)

// BulkDelete deletes multiple records in the same table based on a column value
// partial success possible: if some vals are not found, an error will be returned containing the failed vals, but the other rows will be deleted
func BulkDelete[T any](ctx context.Context, db PoolOrTx, schemaName, tableName, columnName string, vals []T) error {

	if len(vals) == 0 {
		return fmt.Errorf("len(vals) is %v", len(vals))
	}

	stmt := fmt.Sprintf("DELETE FROM %s.%s WHERE %s = $1;", schemaName, tableName, columnName)
	batch := &pgx.Batch{}
	invalidVals := []T{}

	// for each value to be deleted
	for _, v := range vals {

		// queue the query
		batch.Queue(stmt, v).Exec(func(ct pgconn.CommandTag) error {
			if ct.RowsAffected() == 0 {
				invalidVals = append(invalidVals, v)
			}
			return nil
		})
	}

	// send all queries to db
	// any SQL syntax errors will fail here and no rows will be deleted
	err := db.SendBatch(ctx, batch).Close()
	if err != nil {
		return lyserr.Db{Err: fmt.Errorf("db.SendBatch.Close failed: %w", err)}
	}

	// if some vals were invalid, return them
	if len(invalidVals) > 0 {
		rets := make([]string, len(invalidVals))
		for i, v := range invalidVals {
			rets[i] = fmt.Sprintf("%v", v)
		}
		return lyserr.Db{Err: fmt.Errorf("partial success: invalid vals: %s", strings.Join(rets, ", "))}
	}

	return nil
}
