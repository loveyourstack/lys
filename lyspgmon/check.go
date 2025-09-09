package lyspgmon

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CheckDDL checks the DDL integrity of the database. It should be run after schema updates and also periodically
func CheckDDL(ctx context.Context, ownerDb *pgxpool.Pool, infoLog *slog.Logger) (err error) {

	// add any missing updated_at triggers
	err = AddMissingUpdatedAtTriggers(ctx, ownerDb, infoLog)
	if err != nil {
		return fmt.Errorf("AddMissingUpdatedAtTriggers failed: %w", err)
	}

	// check for duplicate shortnames
	// TODO

	// check that _archive tables have all columns of their base table
	// TODO

	return nil
}
