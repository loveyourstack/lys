package lyspgdb

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

func getConnStr(dbConfig Database, userConfig User) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", userConfig.Name, userConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)
}

// GetConn returns a connection pool to the postgres database matching the config params
func GetConfig(dbConfig Database, userConfig User) (cfg *pgxpool.Config, err error) {

	cfg, err = pgxpool.ParseConfig(getConnStr(dbConfig, userConfig))
	if err != nil {
		return nil, fmt.Errorf("pgxpool.ParseConfig failed: %w", err)
	}

	return cfg, nil
}

// GetPool returns a connection pool to the postgres database matching the config params
func GetPool(ctx context.Context, dbConfig Database, userConfig User) (db *pgxpool.Pool, err error) {

	db, err = pgxpool.New(ctx, getConnStr(dbConfig, userConfig))
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New failed: %w", err)
	}

	return db, nil
}

// GetPoolWithTypes returns a connection pool and registers the supplied types to each connection
func GetPoolWithTypes(ctx context.Context, dbConfig Database, userConfig User, dataTypeNames []string) (db *pgxpool.Pool, err error) {

	cfg, err := GetConfig(dbConfig, userConfig)
	if err != nil {
		return nil, fmt.Errorf("GetConfig failed: %w", err)
	}

	// register types in conn AfterConnect hook
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		for _, typeName := range dataTypeNames {
			dataType, err := conn.LoadType(ctx, typeName)
			if err != nil {
				return fmt.Errorf("conn.LoadType failed for type: %s: %w", typeName, err)
			}
			conn.TypeMap().RegisterType(dataType)
		}
		return nil
	}

	db, err = pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.NewWithConfig failed: %w", err)
	}

	return db, nil
}

// executeFile extracts the supplied file from the supplied filesystem and executes it into supplied database
func executeFile(ctx context.Context, db *pgxpool.Pool, sqlFileName string, sqlAssets embed.FS, replacementsMap map[string]string, infoLog *slog.Logger) (err error) {

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
