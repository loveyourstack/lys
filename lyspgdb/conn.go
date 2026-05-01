package lyspgdb

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func getConnStr(dbConfig Database, userConfig User, appName string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?application_name=%s", userConfig.Name, userConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database, appName)
}

// GetConfig returns a Config struct matching the supplied params
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

	err = db.Ping(ctx)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("db.Ping failed: %w", err)
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

	err = db.Ping(ctx)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("db.Ping failed: %w", err)
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

	cfg.PrepareConn = func(ctx context.Context, conn *pgx.Conn) (bool, error) {

		ctxVal, ok := ctx.Value(ctxKey).(ctxValueT)
		if !ok {
			return false, fmt.Errorf("ctxVal not found in ctx")
		}

		// set ctx value into this connection's setting
		_, err := conn.Exec(ctx, fmt.Sprintf("SET %s TO %v;", pgx.Identifier{settingName}.Sanitize(), ctxVal))
		if err != nil {
			return false, fmt.Errorf("conn.Exec (set ctx value) failed: %w", err)
		}

		return true, nil
	}

	cfg.AfterRelease = func(conn *pgx.Conn) bool {

		// unset the setting before this connection is released to pool
		_, err := conn.Exec(context.Background(), fmt.Sprintf("SET %s TO '';", pgx.Identifier{settingName}.Sanitize()))
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

	// don't try to ping here, it causes an infinite loop

	return db, nil
}
