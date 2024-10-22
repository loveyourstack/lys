package lyspg

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/lyserr"
)

type Column struct {
	Name        string `db:"column_name"`
	DataType    string `db:"data_type"`
	IsNullable  bool   `db:"is_nullable"`
	IsIdentity  bool   `db:"is_identity"`
	IsGenerated bool   `db:"is_generated"`
	IsTracking  bool
}

type ForeignKey struct {
	ConstraintName string `db:"constraint_name"`
	ChildSchema    string `db:"child_schema"`
	ChildTable     string `db:"child_table"`
	ChildColumn    string `db:"child_column"`
	ParentSchema   string `db:"parent_schema"`
	ParentTable    string `db:"parent_table"`
	ParentColumn   string `db:"parent_column"`
}

// GetChildForeignKeys returns the child FKs of the supplied table
func GetChildForeignKeys(ctx context.Context, db PoolOrTx, schemaName, tableName string) (fks []ForeignKey, err error) {

	stmt := `SELECT tc.constraint_name, tc.table_schema AS child_schema, tc.table_name AS child_table, kcu.column_name AS child_column, ccu.table_schema AS parent_schema, 
		ccu.table_name AS parent_table, ccu.column_name AS parent_column
	FROM information_schema.table_constraints tc 
	JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name AND tc.table_schema = kcu.table_schema
	JOIN information_schema.constraint_column_usage ccu ON ccu.constraint_name = tc.constraint_name
	WHERE tc.constraint_type = 'FOREIGN KEY' AND ccu.table_schema = $1 AND ccu.table_name = $2;`

	rows, _ := db.Query(ctx, stmt, schemaName, tableName)
	fks, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[ForeignKey])
	if err != nil {
		return nil, lyserr.Db{Err: fmt.Errorf("pgx.CollectRows failed: %w", err), Stmt: stmt}
	}

	return fks, nil
}

// GetForeignKeys returns the parent FKs of the supplied table
// caution: will return incomplete or empty result set if user is not table owner
func GetForeignKeys(ctx context.Context, db PoolOrTx, schemaName, tableName string) (fks []ForeignKey, err error) {

	stmt := `SELECT tc.constraint_name, tc.table_schema AS child_schema, tc.table_name AS child_table, kcu.column_name AS child_column, ccu.table_schema AS parent_schema, 
		ccu.table_name AS parent_table, ccu.column_name AS parent_column
	FROM information_schema.table_constraints tc 
	JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name AND tc.table_schema = kcu.table_schema
	JOIN information_schema.constraint_column_usage ccu ON ccu.constraint_name = tc.constraint_name
	WHERE tc.constraint_type = 'FOREIGN KEY' AND tc.table_schema = $1 AND tc.table_name = $2;`

	rows, _ := db.Query(ctx, stmt, schemaName, tableName)
	fks, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[ForeignKey])
	if err != nil {
		return nil, lyserr.Db{Err: fmt.Errorf("pgx.CollectRows failed: %w", err), Stmt: stmt}
	}

	return fks, nil
}

func GetTableColumns(ctx context.Context, db PoolOrTx, schemaName, tableName string) (cols []Column, err error) {

	stmt := `SELECT column_name, data_type, 
		CASE WHEN is_nullable = 'YES' THEN true ELSE false END AS is_nullable, 
		CASE WHEN is_identity = 'YES' THEN true ELSE false END AS is_identity,
		CASE WHEN is_generated = 'ALWAYS' THEN true ELSE false END AS is_generated
	FROM information_schema.columns 
	WHERE table_schema = $1 AND table_name = $2
	ORDER BY is_identity DESC, column_name;`

	rows, _ := db.Query(ctx, stmt, schemaName, tableName)
	cols, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[Column])
	if err != nil {
		return nil, lyserr.Db{Err: fmt.Errorf("pgx.CollectRows failed: %w", err), Stmt: stmt}
	}

	// assign tracking cols
	for i := range cols {
		if slices.Contains([]string{"entry_at", "entry_by", "last_modified_at", "last_modified_by"}, cols[i].Name) {
			cols[i].IsTracking = true
		}
	}

	return cols, nil
}

func GetStatsRowCount(ctx context.Context, db PoolOrTx, schemaName, tableName string) (rowCount int64, err error) {

	stmt := fmt.Sprintf("SELECT reltuples::bigint FROM pg_class WHERE oid = ('%s.%s')::regclass;", schemaName, tableName)

	rows, _ := db.Query(ctx, stmt)
	rowCount, err = pgx.CollectExactlyOneRow(rows, pgx.RowTo[int64])
	if err != nil {
		return 0, lyserr.Db{Err: fmt.Errorf("pgx.CollectExactlyOneRow failed: %w", err), Stmt: stmt}
	}

	return rowCount, nil
}

func GetTableColumnNames(ctx context.Context, db PoolOrTx, schemaName, tableName string) (colNames []string, err error) {

	stmt := `SELECT column_name FROM information_schema.columns WHERE table_schema = $1 AND table_name = $2;`

	rows, _ := db.Query(ctx, stmt, schemaName, tableName)
	colNames, err = pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, lyserr.Db{Err: fmt.Errorf("pgx.CollectRows failed: %w", err), Stmt: stmt}
	}

	return colNames, nil
}

func GetTableComment(ctx context.Context, db PoolOrTx, schemaName, tableName string) (comment string, err error) {

	stmt := fmt.Sprintf("SELECT obj_description('%s.%s'::regclass);", schemaName, tableName)

	rows, _ := db.Query(ctx, stmt)
	comment, err = pgx.CollectExactlyOneRow(rows, pgx.RowTo[string])
	if err != nil {
		return "", lyserr.Db{Err: fmt.Errorf("pgx.CollectExactlyOneRow failed: %w", err), Stmt: stmt}
	}

	return comment, nil
}

func GetTableShortName(ctx context.Context, db PoolOrTx, schemaName, tableName string) (shortName string, err error) {

	comment, err := GetTableComment(ctx, db, schemaName, tableName)
	if err != nil {
		return "", fmt.Errorf("GetTableComment failed: %w", err)
	}

	snPrefix := "shortname: "

	if len(comment) < len(snPrefix)+1 || !strings.Contains(comment, snPrefix) {
		return "", nil
	}

	return strings.Replace(comment, snPrefix, "", 1), nil
}
