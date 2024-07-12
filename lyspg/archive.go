package lyspg

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// experimental

var (
	ArchivedTableSuffix = "_archived"
	ArchivedTableCols   = []string{"archived_at", "archived_by_cascade"}
)

// Archive moves record(s) to a table's corresponding archived table using the supplied tx
// the archived table name must be the source table name plus ArchivedTableSuffix
// the archived table must have the same columns as the source table and also the ArchivedTableCols defined above
// the column order does not need to be the same in the source and archived tables
// if this func returns an error, rollback the tx
func Archive(ctx context.Context, tx pgx.Tx, schemaName, tableName, pKColName string, pkVal any, isCascaded bool) (stmt string, err error) {

	// get main table cols from info schema
	tableCols, stmt, err := GetTableColumnNames(ctx, tx, schemaName, tableName)
	if err != nil {
		return stmt, fmt.Errorf("GetTableColumnNames failed: %w", err)
	}

	// insert record(s) into archived table
	archivedTableCols := append(tableCols, ArchivedTableCols...)

	stmt = fmt.Sprintf("INSERT INTO %s.%s (%s) SELECT %s, now(), %t FROM %s.%s WHERE %s = $1;",
		schemaName, tableName+ArchivedTableSuffix, strings.Join(archivedTableCols, ","),
		strings.Join(tableCols, ","), isCascaded, schemaName, tableName, pKColName)
	cmdTag, err := tx.Exec(ctx, stmt, pkVal)
	if err != nil {
		return stmt, fmt.Errorf("tx.Exec (Insert) failed: %w", err)
	}

	// if no rows affected, pkVal doesn't exist
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

// Restore moves a previously archived record to the corresponding table using the supplied tx
// if this func returns an error, rollback the tx
func Restore(ctx context.Context, tx pgx.Tx, schemaName, tableName, pkColName string, pkVal any, isCascaded bool) (stmt string, err error) {

	// get main table cols from info schema
	tableCols, stmt, err := GetTableColumnNames(ctx, tx, schemaName, tableName)
	if err != nil {
		return stmt, fmt.Errorf("GetTableColumnNames failed: %w", err)
	}

	// insert record(s) into main table
	stmt = fmt.Sprintf("INSERT INTO %s.%s (%s) SELECT %s FROM %s.%s WHERE archived_by_cascade = %t AND %s = $1;",
		schemaName, tableName, strings.Join(tableCols, ","),
		strings.Join(tableCols, ","), schemaName, tableName+ArchivedTableSuffix, isCascaded, pkColName)
	cmdTag, err := tx.Exec(ctx, stmt, pkVal)
	if err != nil {
		return stmt, fmt.Errorf("tx.Exec (Insert) failed: %w", err)
	}

	// if no rows affected, Id doesn't exist
	if cmdTag.RowsAffected() == 0 {
		return "", pgx.ErrNoRows
	}

	// delete record(s) from archived table
	stmt = fmt.Sprintf("DELETE FROM %s.%s WHERE archived_by_cascade = %t AND %s = $1;", schemaName, tableName+ArchivedTableSuffix, isCascaded, pkColName)
	_, err = tx.Exec(ctx, stmt, pkVal)
	if err != nil {
		return stmt, fmt.Errorf("tx.Exec (Delete) failed: %w", err)
	}

	return "", nil
}
