package lyspgmon

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lyserr"
)

// CheckDDL checks the DDL integrity of the database. It should be run after schema updates and also periodically
func CheckDDL(ctx context.Context, ownerDb *pgxpool.Pool, infoLog, errorLog *slog.Logger) (err error) {

	// add any missing updated_at triggers
	err = AddMissingUpdatedAtTriggers(ctx, ownerDb, infoLog)
	if err != nil {
		return fmt.Errorf("AddMissingUpdatedAtTriggers failed: %w", err)
	}

	// check for tables that have the t_audit_update trigger but are missing the last_user_update_by col
	err = CheckMissingLastUserUpdateByCols(ctx, ownerDb, errorLog)
	if err != nil {
		return fmt.Errorf("CheckMissingLastUserUpdateByCols failed: %w", err)
	}

	// check for duplicate table shortnames
	err = CheckDuplicateShortnames(ctx, ownerDb, errorLog)
	if err != nil {
		return fmt.Errorf("CheckDuplicateShortnames failed: %w", err)
	}

	// check that _archived tables columns are consistent with their base table
	err = CheckInconsistentArchivedCols(ctx, ownerDb, errorLog)
	if err != nil {
		return fmt.Errorf("CheckInconsistentArchivedCols failed: %w", err)
	}

	return nil
}

// AddMissingUpdatedAtTriggers adds missing updated_at triggers for all tables returned by v_missing_updated_at_trigger
func AddMissingUpdatedAtTriggers(ctx context.Context, ownerDb *pgxpool.Pool, infoLog *slog.Logger) (err error) {

	type missingTrigger struct {
		TableSchema string `db:"table_schema"`
		TableName   string `db:"table_name"`
	}

	// select tables missing the trigger
	stmt := "SELECT table_schema, table_name FROM lyspgmon.v_missing_updated_at_trigger;"
	rows, _ := ownerDb.Query(ctx, stmt)
	items, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[missingTrigger])
	if err != nil {
		return lyserr.Db{Err: fmt.Errorf("pgx.CollectRows failed: %w", err), Stmt: stmt}
	}

	// exit if none found
	if len(items) == 0 {
		return nil
	}

	// for each table
	for _, item := range items {

		// create the trigger
		stmt = fmt.Sprintf("CREATE TRIGGER t_set_updated_at BEFORE UPDATE ON %s.%s FOR EACH ROW EXECUTE PROCEDURE system.set_updated_at();",
			item.TableSchema, item.TableName)
		_, err = ownerDb.Exec(ctx, stmt)
		if err != nil {
			return lyserr.Db{Err: fmt.Errorf("ownerDb.Exec failed on %s.%s: %w", item.TableSchema, item.TableName, err), Stmt: stmt}
		}

		infoLog.Info("created set_updated_at trigger", slog.String("schema", item.TableSchema), slog.String("table", item.TableName))
	}

	return nil
}

// CheckDuplicateShortnames checks for tables with the same shortname comment
func CheckDuplicateShortnames(ctx context.Context, ownerDb *pgxpool.Pool, errorLog *slog.Logger) (err error) {

	type dupShortname struct {
		Comment     string `db:"com"`
		TableSchema string `db:"table_schema"`
		TableName   string `db:"table_name"`
	}

	// select tables with duplicate shortname comments
	stmt := "SELECT com, table_schema, table_name FROM lyspgmon.v_duplicate_shortnames;"
	rows, _ := ownerDb.Query(ctx, stmt)
	items, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[dupShortname])
	if err != nil {
		return lyserr.Db{Err: fmt.Errorf("pgx.CollectRows failed: %w", err), Stmt: stmt}
	}

	// exit if none found
	if len(items) == 0 {
		return nil
	}

	// list the dups
	for _, item := range items {

		errorLog.Warn("has duplicate shortname", slog.String("comment", item.Comment), slog.String("schema", item.TableSchema), slog.String("table", item.TableName))
	}

	return nil
}

// CheckInconsistentArchivedCols checks for archived tables that do not have the same cols as their base tables
func CheckInconsistentArchivedCols(ctx context.Context, ownerDb *pgxpool.Pool, errorLog *slog.Logger) (err error) {

	type dupShortname struct {
		Info        string `db:"info"`
		TableSchema string `db:"table_schema"`
		TableName   string `db:"table_name"`
		ColumnName  string `db:"column_name"`
	}

	// select tables with inconsistent archived cols
	stmt := "SELECT info, table_schema, table_name, column_name FROM lyspgmon.v_inconsistent_archived_cols;"
	rows, _ := ownerDb.Query(ctx, stmt)
	items, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[dupShortname])
	if err != nil {
		return lyserr.Db{Err: fmt.Errorf("pgx.CollectRows failed: %w", err), Stmt: stmt}
	}

	// exit if none found
	if len(items) == 0 {
		return nil
	}

	// list the inconsistencies
	for _, item := range items {

		errorLog.Warn("inconsistent archived cols", slog.String("info", item.Info), slog.String("schema", item.TableSchema), slog.String("table", item.TableName),
			slog.String("column", item.ColumnName))
	}

	return nil
}

// CheckMissingLastUserUpdateByCols checks for tables that have the t_audit_update trigger but are missing the last_user_update_by col
func CheckMissingLastUserUpdateByCols(ctx context.Context, ownerDb *pgxpool.Pool, errorLog *slog.Logger) (err error) {

	type missingCol struct {
		EventObjectSchema string `db:"event_object_schema"`
		EventObjectTable  string `db:"event_object_table"`
	}

	// select tables missing the col
	stmt := "SELECT event_object_schema, event_object_table FROM lyspgmon.v_missing_last_user_update_by_col;"
	rows, _ := ownerDb.Query(ctx, stmt)
	items, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[missingCol])
	if err != nil {
		return lyserr.Db{Err: fmt.Errorf("pgx.CollectRows failed: %w", err), Stmt: stmt}
	}

	// exit if none found
	if len(items) == 0 {
		return nil
	}

	// for each col
	for _, item := range items {

		// report it missing
		errorLog.Warn("has audit trigger but missing last_user_update_by col", slog.String("schema", item.EventObjectSchema), slog.String("table", item.EventObjectTable))
	}

	return nil
}
