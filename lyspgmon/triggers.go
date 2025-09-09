package lyspgmon

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lyserr"
)

// AddMissingUpdatedAtTriggers adds missing updated_at triggers for all tables returned by v_missing_updated_at_trigger
func AddMissingUpdatedAtTriggers(ctx context.Context, ownerDb *pgxpool.Pool, infoLog *slog.Logger) (err error) {

	type missingTrigger struct {
		TableSchema string `db:"table_schema"`
		TableName   string `db:"table_name"`
		Event       string `db:"event"`
	}

	// select missing triggers
	stmt := "SELECT table_schema, table_name, event FROM lyspgmon.v_missing_updated_at_trigger"
	rows, _ := ownerDb.Query(ctx, stmt)
	items, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[missingTrigger])
	if err != nil {
		return lyserr.Db{Err: fmt.Errorf("pgx.CollectRows failed: %w", err), Stmt: stmt}
	}

	// exit if none found
	if len(items) == 0 {
		return nil
	}

	// for each missing trigger
	for _, item := range items {

		suffix := ""
		switch item.Event {
		case "INSERT":
			suffix = "_i"
		case "UPDATE":
			suffix = "_u"
		default:
			return fmt.Errorf("unknown event: %s", item.Event)
		}

		stmt = fmt.Sprintf("CREATE TRIGGER %s_set_updated_at%s BEFORE %s ON %s.%s FOR EACH ROW EXECUTE PROCEDURE system.set_updated_at();",
			item.TableName, suffix, item.Event, item.TableSchema, item.TableName)
		_, err = ownerDb.Exec(ctx, stmt)
		if err != nil {
			return lyserr.Db{Err: fmt.Errorf("ownerDb.Exec (create trigger) failed on %s.%s: %w", item.TableSchema, item.TableName, err), Stmt: stmt}
		}

		infoLog.Info("created set_updated_at trigger", slog.String("schema", item.TableSchema), slog.String("table", item.TableName), slog.String("event", item.Event))
	}

	return nil
}
