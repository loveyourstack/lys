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

// FileReplacement holds text replacement details for ExecuteFile.
// Use when placeholders such as usernames or passwords need to be injected at runtime.
type FileReplacement struct {
	From string
	To   string
}

// ExecuteFile extracts the supplied file from the supplied filesystem and executes it into supplied database
func ExecuteFile(ctx context.Context, db *pgxpool.Pool, sqlFileName string, sqlAssets embed.FS, replacements []FileReplacement, infoLog *slog.Logger) (err error) {

	// get file content from packaged sql
	rawQry, err := fs.ReadFile(sqlAssets, sqlFileName)
	if err != nil {
		return fmt.Errorf("fs.ReadFile failed for file: %v: %w", sqlFileName, err)
	}

	infoLog.Info("Executing " + sqlFileName)

	// make text replacements, if any
	for _, r := range replacements {
		rawQry = bytes.ReplaceAll(rawQry, []byte(r.From), []byte(r.To))
	}

	// execute file content in the db
	if _, err = db.Exec(ctx, string(rawQry)); err != nil {
		return fmt.Errorf("db.Exec failed for file: %v: %w", sqlFileName, err)
	}

	return nil
}
