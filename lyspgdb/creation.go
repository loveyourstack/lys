package lyspgdb

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateLocalDb creates or recreates a test or dev db
func CreateLocalDb(ctx context.Context, sqlAssets embed.FS, dbConf Database, dbSuperUser, dbOwnerConf User,
	dropExisting, addSecurityPermissions bool, replacements []FileReplacement, logger *slog.Logger) (err error) {

	pgDbConf := Database{
		Host:     dbConf.Host,
		Port:     dbConf.Port,
		Database: "postgres",
	}

	pgSuperUserConf := User{
		Name:     dbSuperUser.Name,
		Password: dbSuperUser.Password,
	}

	// connect with superuser to postgres db
	pgSuperUserPgDb, err := GetPool(ctx, pgDbConf, pgSuperUserConf, "test")
	if err != nil {
		return fmt.Errorf("GetPool failed (postgres db with %v user): %w", dbSuperUser.Name, err)
	}
	defer pgSuperUserPgDb.Close()

	// (re-)create database
	if dropExisting {
		logger.Info("Dropping database " + dbConf.Database + " if exists")
		if err = DropDb(ctx, pgSuperUserPgDb, dbConf.Database); err != nil {
			return fmt.Errorf("DropDb failed for database: %v: %w", dbConf.Database, err)
		}
	}

	logger.Info("Creating database " + dbConf.Database)
	if err = CreateDb(ctx, pgSuperUserPgDb, dbConf.Database); err != nil {
		return fmt.Errorf("CreateDb failed for database: %v: %w", dbConf.Database, err)
	}

	// ----------------

	// connect with superuser user to target db
	pgSuperUserDb, err := GetPool(ctx, dbConf, pgSuperUserConf, "CreateLocalDb func")
	if err != nil {
		return fmt.Errorf("GetPool failed (database %v with %v user): %w", dbConf.Database, dbSuperUser.Name, err)
	}
	defer pgSuperUserDb.Close()

	// add database extensions
	logger.Info("Adding extensions, if any")
	if err = ExecuteFile(ctx, pgSuperUserDb, "extensions.sql", sqlAssets, replacements, logger); err != nil {
		return fmt.Errorf("ExecuteFile failed (extensions): %w", err)
	}

	// grant all rights on the db to the db owner user
	logger.Info("Granting all rights to " + dbOwnerConf.Name)
	if err = GrantAll(ctx, pgSuperUserDb, dbOwnerConf, dbConf.Database); err != nil {
		return fmt.Errorf("GrantAll failed for all rights for user: %v: %w", dbOwnerConf.Name, err)
	}

	// add other security permissions from file
	if addSecurityPermissions {
		logger.Info("Adding security permissions")
		if err = ExecuteFile(ctx, pgSuperUserDb, "security_permissions.sql", sqlAssets, replacements, logger); err != nil {
			return fmt.Errorf("ExecuteFile failed (security_permissions): %w", err)
		}
	}

	// ----------------

	// connect with db owner user to target db
	dbOwnerUserDb, err := GetPool(ctx, dbConf, dbOwnerConf, "CreateLocalDb func")
	if err != nil {
		return fmt.Errorf("GetPool failed (database %v with %v user): %w", dbConf.Database, dbOwnerConf.Name, err)
	}
	defer dbOwnerUserDb.Close()

	// populate and analyze db

	logger.Info("Populating database")
	if err = PopulateDb(ctx, dbOwnerUserDb, sqlAssets, dbConf.SchemaCreationOrder, replacements, logger); err != nil {
		return fmt.Errorf("PopulateDb failed: %w", err)
	}

	logger.Info("Analyze")
	if _, err = dbOwnerUserDb.Exec(ctx, "ANALYZE;"); err != nil {
		return fmt.Errorf("dbOwnerUserDb.Exec failed (ANALYZE): %w", err)
	}

	return nil
}

// DropDb deletes a database
// pgUserPgDb is a connection to the postgres database with the postgres user
func DropDb(ctx context.Context, pgUserPgDb *pgxpool.Pool, dbName string) (err error) {

	// drop the database if needed
	stmt := fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE);", pgx.Identifier{dbName}.Sanitize())
	if _, err = pgUserPgDb.Exec(ctx, stmt); err != nil {
		return fmt.Errorf("pgUserPgDb.Exec failed: %w", err)
	}

	return nil
}

// CreateDb creates a database
// pgUserPgDb is a connection to the postgres database with the postgres user
func CreateDb(ctx context.Context, pgUserPgDb *pgxpool.Pool, dbName string) (err error) {

	// create the database
	stmt := fmt.Sprintf("CREATE DATABASE %s;", pgx.Identifier{dbName}.Sanitize())
	if _, err = pgUserPgDb.Exec(ctx, stmt); err != nil {
		return fmt.Errorf("pgUserPgDb.Exec failed: %w", err)
	}

	return nil
}

// GrantAll grants all rights to the specified user on the specified db
// pgUserDb is a connection to the target database with the postgres user
func GrantAll(ctx context.Context, pgUserDb *pgxpool.Pool, userConf User, dbName string) (err error) {

	// grant all rights on this database
	stmt := fmt.Sprintf("GRANT ALL ON DATABASE %s TO %s;", pgx.Identifier{dbName}.Sanitize(), pgx.Identifier{userConf.Name}.Sanitize())
	if _, err = pgUserDb.Exec(ctx, stmt); err != nil {
		return fmt.Errorf("pgUserDb.Exec (grant all on database) failed: %w", err)
	}

	// grant access to public schema
	stmt = fmt.Sprintf("GRANT ALL ON SCHEMA public TO %s;", pgx.Identifier{userConf.Name}.Sanitize())
	if _, err = pgUserDb.Exec(ctx, stmt); err != nil {
		return fmt.Errorf("pgUserDb.Exec (grant all on schema public) failed: %w", err)
	}

	return nil
}

// PopulateDb writes schema, tables, functions and views
func PopulateDb(ctx context.Context, db *pgxpool.Pool, sqlAssets embed.FS, schemaCreationOrder []string, replacements []FileReplacement,
	logger *slog.Logger) (err error) {

	// make sure db is empty (no user objects)
	stmt := `SELECT count(*) FROM pg_class c
	  JOIN pg_namespace s ON s.oid = c.relnamespace
	  WHERE s.nspname NOT IN ('pg_catalog', 'information_schema')
	  AND s.nspname NOT LIKE 'pg_temp%' AND c.relname NOT LIKE 'pg_%'`

	var count int
	row := db.QueryRow(ctx, stmt)
	if err = row.Scan(&count); err != nil {
		return fmt.Errorf("row.Scan failed: %w", err)
	}
	if count != 0 {
		return fmt.Errorf("target database is not empty, it contains %d user objects", count)
	}

	// create schemas and assign default permissions
	if err = ExecuteFile(ctx, db, "schemas.sql", sqlAssets, replacements, logger); err != nil {
		return fmt.Errorf("ExecuteFile failed (schemas): %w", err)
	}

	// add assets (in order) which should be added before functions
	assetTypes := []string{"types", "domains", "sequences", "tables"}
	for _, assetType := range assetTypes {
		for _, schema := range schemaCreationOrder {
			assetPath := fmt.Sprintf("%s/%s_%s.sql", schema, schema, assetType)
			if err = ExecuteFile(ctx, db, assetPath, sqlAssets, replacements, logger); err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					continue
				}
				return fmt.Errorf("ExecuteFile failed for schema: %v, asset type: %v: %w", schema, assetType, err)
			}
		}
	}

	// funcs and views: can interdepend, so process together
	// note: they are added in schemaCreationOrder. Any inter-schema dependencies must point backwards, not forwards
	if err = addFuncsAndViews(ctx, db, sqlAssets, schemaCreationOrder, replacements, logger); err != nil {
		return fmt.Errorf("addFuncsAndViews failed: %w", err)
	}

	// add live then test data, if any
	assetTypes = []string{"data", "test_data"}
	for _, assetType := range assetTypes {
		for _, schema := range schemaCreationOrder {
			assetPath := fmt.Sprintf("%s/%s_%s.sql", schema, schema, assetType)
			if err = ExecuteFile(ctx, db, assetPath, sqlAssets, replacements, logger); err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					continue
				}
				return fmt.Errorf("ExecuteFile failed for schema: %v, asset type: %v: %w", schema, assetType, err)
			}
		}
	}

	// add trigger func assignments to tables
	assetTypes = []string{"tfa_"}
	for _, assetType := range assetTypes {
		for _, schema := range schemaCreationOrder {
			dirEntries, err := sqlAssets.ReadDir("" + schema)
			if err != nil {
				return fmt.Errorf("sqlAssets.ReadDir failed for schema: %v: %w", schema, err)
			}
			for _, dirEntry := range dirEntries {
				if strings.HasPrefix(dirEntry.Name(), assetType) {
					if err = ExecuteFile(ctx, db, schema+"/"+dirEntry.Name(), sqlAssets, replacements, logger); err != nil {
						return fmt.Errorf("ExecuteFile failed for schema: %v, asset type: %v: %w", schema, assetType, err)
					}
				}
			}
		}
	}

	return nil
}

// addFuncsAndViews attempts to add all functions and views for each schema, handling interdependencies by deferring execution of files that fail due to missing dependencies.
func addFuncsAndViews(ctx context.Context, db *pgxpool.Pool, sqlAssets embed.FS, schemaCreationOrder []string, replacements []FileReplacement,
	logger *slog.Logger) error {

	for _, schema := range schemaCreationOrder {

		filePaths := []string{}

		// get files from /funcs, if any
		funcFilePaths, err := getSqlFilePathsfromDir(sqlAssets, schema+"/funcs")
		if err != nil {
			return fmt.Errorf("getSqlFilePathsfromDir failed for schema: %v, funcs: %w", schema, err)
		}
		filePaths = append(filePaths, funcFilePaths...)

		// get files from /views, if any
		viewFilePaths, err := getSqlFilePathsfromDir(sqlAssets, schema+"/views")
		if err != nil {
			return fmt.Errorf("getSqlFilePathsfromDir failed for schema: %v, views: %w", schema, err)
		}
		filePaths = append(filePaths, viewFilePaths...)

		// loop until all files are executed or no progress can be made
		pending := filePaths
		for len(pending) > 0 {

			var deferred []string
			progress := false

			// attempt to execute each pending file, deferring those that fail due to missing dependencies
			for _, f := range pending {

				err := ExecuteFile(ctx, db, f, sqlAssets, replacements, logger)
				if err == nil {
					progress = true
					continue
				}

				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) &&
					(pgErr.Code == pgerrcode.UndefinedTable || pgErr.Code == pgerrcode.UndefinedFunction) {
					deferred = append(deferred, f) // depends on a func/view not yet created
					continue
				}

				return err // genuine error: fail fast

			} // next pending file

			if !progress {
				return fmt.Errorf("unresolvable dependencies: %v", deferred)
			}

			pending = deferred

		} // next attempt

	} // next schema

	return nil
}

// getFilePathsfromDir retrieves all file paths from the specified directory within the embedded filesystem.
// If the directory does not exist, it returns an empty slice without error.
func getSqlFilePathsfromDir(sqlAssets embed.FS, dir string) ([]string, error) {

	filePaths := []string{}

	dirEntries, err := sqlAssets.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return filePaths, nil
		}
		return nil, fmt.Errorf("sqlAssets.ReadDir failed for dir: %v: %w", dir, err)
	}

	for _, dirEntry := range dirEntries {

		// skip dirs or non SQL files (e.g. to allow READMEs)
		if dirEntry.IsDir() || !strings.HasSuffix(dirEntry.Name(), ".sql") {
			continue
		}

		filePaths = append(filePaths, dir+"/"+dirEntry.Name())
	}

	return filePaths, nil
}
