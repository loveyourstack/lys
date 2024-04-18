package lyspg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DeleteByValue deletes from the supplied database table using the value of any column
// must delete at least 1 row
func DeleteByValue(ctx context.Context, db PoolOrTx, schemaName, tableName, columnName string, val any) (stmt string, err error) {

	stmt = fmt.Sprintf("DELETE FROM %s.%s WHERE %s = $1;", schemaName, tableName, columnName)

	res, err := db.Exec(ctx, stmt, val)
	if err != nil {
		return stmt, fmt.Errorf("db.Exec failed: %w", err)
	}

	if res.RowsAffected() == 0 {
		return stmt, pgx.ErrNoRows
	}

	return "", nil
}

// DeleteUnique is guaranteed to delete a single row from the supplied database table using the unique value supplied
func DeleteUnique(ctx context.Context, db *pgxpool.Pool, schemaName, tableName, columnName string, uniqueVal any) (stmt string, err error) {

	// begin tx
	tx, err := db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("db.Begin failed: %w", err)
	}
	defer tx.Rollback(ctx)

	stmt = fmt.Sprintf("DELETE FROM %s.%s WHERE %s = $1;", schemaName, tableName, columnName)

	res, err := tx.Exec(ctx, stmt, uniqueVal)
	if err != nil {
		return stmt, fmt.Errorf("tx.Exec failed: %w", err)
	}

	// ensure 1 row was deleted
	if res.RowsAffected() == 0 {
		return stmt, pgx.ErrNoRows
	}
	if res.RowsAffected() > 1 {
		return stmt, pgx.ErrTooManyRows
	}

	// success: commit tx
	err = tx.Commit(ctx)
	if err != nil {
		return "", fmt.Errorf("tx.Commit failed: %w", err)
	}

	return "", nil
}
