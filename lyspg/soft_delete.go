package lyspg

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// experimental

var (
	DeletedTableSuffix = "_deleted"
	DeletedTableCols   = []string{"deleted_at", "deleted_by_cascade"}
)

// SoftDelete moves record(s) to the corresponding deleted table using the supplied tx
// the deleted table must have the same columns as the source table and also the DeletedTableCols defined above
// the deleted table name must be the source table name plus DeletedTableSuffix
func SoftDelete(ctx context.Context, tx pgx.Tx, schemaName, tableName, pKColName string, pkVal any, isCascaded bool) (stmt string, err error) {

	// get main table cols from info schema
	tableCols, stmt, err := GetTableColumnNames(ctx, tx, schemaName, tableName)
	if err != nil {
		return stmt, fmt.Errorf("GetTableColumnNames failed: %w", err)
	}

	// insert record(s) into deleted table
	deletedTableCols := append(tableCols, DeletedTableCols...)

	stmt = fmt.Sprintf("INSERT INTO %s.%s (%s) SELECT %s, now(), %t FROM %s.%s WHERE %s = $1;",
		schemaName, tableName+DeletedTableSuffix, strings.Join(deletedTableCols, ","),
		strings.Join(tableCols, ","), isCascaded, schemaName, tableName, pKColName)
	cmdTag, err := tx.Exec(ctx, stmt, pkVal)
	if err != nil {
		return stmt, fmt.Errorf("tx.Exec (Insert) failed: %w", err)
	}

	// if no rows affected, Id doesn't exist
	if cmdTag.RowsAffected() == 0 {
		return "", pgx.ErrNoRows
	}

	// delete record(s) from main table
	stmt = fmt.Sprintf("DELETE FROM %s.%s WHERE %s = $1;", schemaName, tableName, pKColName)
	_, err = tx.Exec(ctx, stmt, pkVal)
	if err != nil {
		return stmt, fmt.Errorf("tx.Exec (Delete) failed: %w", err)
	}

	return "", nil
}

// Restore moves a previously soft deleted record to the corresponding table using the supplied tx
// if this func returns an error, rollback the tx
func Restore(ctx context.Context, tx pgx.Tx, schemaName, tableName, pkColName string, pkVal any, isCascaded bool) (stmt string, err error) {

	// get main table cols from info schema
	tableCols, stmt, err := GetTableColumnNames(ctx, tx, schemaName, tableName)
	if err != nil {
		return stmt, fmt.Errorf("GetTableColumnNames failed: %w", err)
	}

	// insert record(s) into main table
	stmt = fmt.Sprintf("INSERT INTO %s.%s (%s) SELECT %s FROM %s.%s WHERE deleted_by_cascade = %t AND %s = $1;",
		schemaName, tableName, strings.Join(tableCols, ","),
		strings.Join(tableCols, ","), schemaName, tableName+DeletedTableSuffix, isCascaded, pkColName)
	cmdTag, err := tx.Exec(ctx, stmt, pkVal)
	if err != nil {
		return stmt, fmt.Errorf("tx.Exec (Insert) failed: %w", err)
	}

	// if no rows affected, Id doesn't exist
	if cmdTag.RowsAffected() == 0 {
		return "", pgx.ErrNoRows
	}

	// delete record(s) from deleted table
	stmt = fmt.Sprintf("DELETE FROM %s.%s WHERE deleted_by_cascade = %t AND %s = $1;", schemaName, tableName+DeletedTableSuffix, isCascaded, pkColName)
	_, err = tx.Exec(ctx, stmt, pkVal)
	if err != nil {
		return stmt, fmt.Errorf("tx.Exec (Delete) failed: %w", err)
	}

	return "", nil
}
