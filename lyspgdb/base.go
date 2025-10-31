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

func getConnStr(dbConfig Database, userConfig User, appName string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?application_name=%s", userConfig.Name, userConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database, appName)
}

// GetConn returns a connection pool to the postgres database matching the config params
func GetConfig(dbConfig Database, userConfig User, appName string) (cfg *pgxpool.Config, err error) {

	cfg, err = pgxpool.ParseConfig(getConnStr(dbConfig, userConfig, appName))
	if err != nil {
		return nil, fmt.Errorf("pgxpool.ParseConfig failed: %w", err)
	}

	return cfg, nil
}

// GetPool returns a connection pool to the postgres database matching the config params
func GetPool(ctx context.Context, dbConfig Database, userConfig User, appName string) (db *pgxpool.Pool, err error) {

	db, err = pgxpool.New(ctx, getConnStr(dbConfig, userConfig, appName))
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New failed: %w", err)
	}

	return db, nil
}

// GetPoolWithTypes returns a connection pool and registers the supplied types to each connection
func GetPoolWithTypes(ctx context.Context, dbConfig Database, userConfig User, appName string, dataTypeNames []string) (db *pgxpool.Pool, err error) {

	cfg, err := GetConfig(dbConfig, userConfig, appName)
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

type ContextKey string

// GetPoolWithCtxSetting returns a connection pool wherein each connection has a setting from ctx applied to it on acquisition
// adapted from https://github.com/jackc/pgx/issues/288
func GetPoolWithCtxSetting[ctxValueT any](ctx context.Context, dbConfig Database, userConfig User, appName, settingName string, ctxKey ContextKey, errorLog *slog.Logger) (db *pgxpool.Pool, err error) {

	cfg, err := GetConfig(dbConfig, userConfig, appName)
	if err != nil {
		return nil, fmt.Errorf("GetConfig failed: %w", err)
	}

	cfg.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {

		ctxVal, ok := ctx.Value(ctxKey).(ctxValueT)
		if !ok {
			errorLog.Error("ctxVal not found in ctx")
			return false
		}

		// set ctx value into this connection's setting
		_, err := conn.Exec(ctx, fmt.Sprintf("SET %s TO %v;", settingName, ctxVal))
		if err != nil {
			errorLog.Error("conn.Exec (set ctx value) failed: " + err.Error())
			return false
		}

		return true
	}

	cfg.AfterRelease = func(conn *pgx.Conn) bool {

		// unset the setting before this connection is released to pool
		_, err := conn.Exec(ctx, fmt.Sprintf("SET %s TO '';", settingName))
		if err != nil {
			errorLog.Error("conn.Exec (unset ctx value) failed: " + err.Error())
			return false
		}

		return true
	}

	db, err = pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.NewWithConfig failed: %w", err)
	}

	return db, nil
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
