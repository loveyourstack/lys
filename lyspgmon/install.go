package lyspgmon

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lyspgdb"
	"github.com/loveyourstack/lys/lyspgmon/lyspgmonddl"
)

const schemaName string = "lyspgmon"

// Install creates the lyspgmon schema in the database if it is not already present, and (re)-adds the monitoring views in the lyspgmonddl folder
func Install(ctx context.Context, ownerDb *pgxpool.Pool, dbOwner string, infoLog *slog.Logger) (err error) {

	// create schema if needed
	err = createSchema(ctx, ownerDb, dbOwner)
	if err != nil {
		return fmt.Errorf("createSchema failed: %w", err)
	}

	// execute all embedded views into db
	err = fs.WalkDir(lyspgmonddl.SQLAssets, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("unknown file err: %w", err)
		}

		// skip non-sql files
		if !strings.HasSuffix(d.Name(), ".sql") {
			return nil
		}

		// exec file into db
		err = lyspgdb.ExecuteFile(ctx, ownerDb, d.Name(), lyspgmonddl.SQLAssets, nil, infoLog)
		if err != nil {
			return fmt.Errorf("lyspgdb.ExecuteFile failed for file '%s': %w", d.Name(), err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("fs.WalkDir failed: %w", err)
	}

	return nil
}

func createSchema(ctx context.Context, ownerDb *pgxpool.Pool, dbOwner string) (err error) {

	stmt := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s AUTHORIZATION %s;", schemaName, dbOwner)
	_, err = ownerDb.Exec(ctx, stmt)
	if err != nil {
		return fmt.Errorf("ownerDb.Exec failed: %w", err)
	}

	return nil
}
