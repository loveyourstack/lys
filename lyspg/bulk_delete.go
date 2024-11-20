package lyspg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/lyserr"
)

// BulkDelete deletes multiple records in the same table based on a column value
func BulkDelete[T any](ctx context.Context, db PoolOrTx, schemaName, tableName, columnName string, vals []T) (rowsAffected int64, err error) {

	if len(vals) == 0 {
		return 0, fmt.Errorf("len(vals) is %v", len(vals))
	}

	stmt := fmt.Sprintf("DELETE FROM %s.%s WHERE %s = $1;", schemaName, tableName, columnName)
	batch := &pgx.Batch{}

	// for each value to be deleted
	for _, v := range vals {

		// queue the query
		batch.Queue(stmt, v)
	}

	// send all queries to db
	batchRes := db.SendBatch(ctx, batch)
	defer batchRes.Close()

	// exec all queries
	cmdTag, err := batchRes.Exec()
	if err != nil {
		return 0, lyserr.Db{Err: fmt.Errorf("batchRes.Exec (delete %v recs) failed: %w", batch.Len(), err)}
	}

	if cmdTag.RowsAffected() == 0 {
		return 0, pgx.ErrNoRows
	}

	return cmdTag.RowsAffected(), nil
}
