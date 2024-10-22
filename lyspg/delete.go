package lyspg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lyserr"
)

// DeleteByValue deletes from the supplied database table using the value of any column
// must delete at least 1 row
func DeleteByValue(ctx context.Context, db PoolOrTx, schemaName, tableName, columnName string, val any) error {

	stmt := fmt.Sprintf("DELETE FROM %s.%s WHERE %s = $1;", schemaName, tableName, columnName)

	res, err := db.Exec(ctx, stmt, val)
	if err != nil {
		return lyserr.Db{Err: fmt.Errorf("db.Exec failed: %w", err), Stmt: stmt}
	}

	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// DeleteUnique is guaranteed to delete a single row from the supplied database table using the unique value supplied
func DeleteUnique(ctx context.Context, db *pgxpool.Pool, schemaName, tableName, columnName string, uniqueVal any) error {

	// begin tx
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("db.Begin failed: %w", err)
	}
	defer tx.Rollback(ctx)

	stmt := fmt.Sprintf("DELETE FROM %s.%s WHERE %s = $1;", schemaName, tableName, columnName)

	res, err := tx.Exec(ctx, stmt, uniqueVal)
	if err != nil {
		return lyserr.Db{Err: fmt.Errorf("tx.Exec failed: %w", err), Stmt: stmt}
	}

	// ensure 1 row was deleted
	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	if res.RowsAffected() > 1 {
		return pgx.ErrTooManyRows
	}

	// success: commit tx
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("tx.Commit failed: %w", err)
	}

	return nil
}
