package lyspgdb

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	DbSizePrettyStmt = "SELECT pg_size_pretty(pg_database_size(current_database()));"
	VersionStmt      = "SELECT (string_to_array(version(), ', compiled'))[1];" // omits compilation details
)

// Database holds postgres database config options
type Database struct {
	Host                string
	Port                string
	Database            string
	SchemaCreationOrder []string
}

// User holds postgres user details
type User struct {
	Name     string `toml:"userName"`
	Password string
}

// ExecuteFile extracts the supplied file from the supplied filesystem and executes it into supplied database
func ExecuteFile(ctx context.Context, db *pgxpool.Pool, sqlFileName string, sqlAssets embed.FS, replacementsMap map[string]string, infoLog *slog.Logger) (err error) {

	infoLog.Info("Executing " + sqlFileName)

	// get file content from packaged sql
	rawQry, err := fs.ReadFile(sqlAssets, sqlFileName)
	if err != nil {
		return fmt.Errorf("fs.ReadFile failed for file: %v: %w", sqlFileName, err)
	}

	// make text replacements, if any
	for from, to := range replacementsMap {
		rawQry = bytes.ReplaceAll(rawQry, []byte(from), []byte(to))
	}

	// execute file content in the db
	if _, err = db.Exec(ctx, string(rawQry)); err != nil {
		return fmt.Errorf("db.Exec failed for file: %v: %w", sqlFileName, err)
	}

	return nil
}
