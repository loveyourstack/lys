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

	// check for duplicate shortnames
	// TODO

	// check that _archive tables have all columns of their base table
	// TODO

	return nil
}

// AddMissingUpdatedAtTriggers adds missing updated_at triggers for all tables returned by v_missing_updated_at_trigger
func AddMissingUpdatedAtTriggers(ctx context.Context, ownerDb *pgxpool.Pool, infoLog *slog.Logger) (err error) {

	type missingTrigger struct {
		TableSchema string `db:"table_schema"`
		TableName   string `db:"table_name"`
	}

	// select tables missing the trigger
	stmt := "SELECT table_schema, table_name FROM lyspgmon.v_missing_updated_at_trigger"
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

// CheckMissingLastUserUpdateByCols checks for tables that have the t_audit_update trigger but are missing the last_user_update_by col
func CheckMissingLastUserUpdateByCols(ctx context.Context, ownerDb *pgxpool.Pool, errorLog *slog.Logger) (err error) {

	type missingCol struct {
		EventObjectSchema string `db:"event_object_schema"`
		EventObjectTable  string `db:"event_object_table"`
	}

	// select tables missing the col
	stmt := "SELECT event_object_schema, event_object_table FROM lyspgmon.v_missing_last_user_update_by_col"
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
