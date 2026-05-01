package lyspgdb

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateLocalDb creates or recreates a test or dev db
func CreateLocalDb(ctx context.Context, sqlAssets embed.FS, dbConf Database, dbSuperUser, dbOwnerConf User,
	dropExisting, addSecurityPermissions bool, replacements []FileReplacement, infoLog *slog.Logger) (err error) {

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
		infoLog.Info("Dropping database " + dbConf.Database + " if exists")
		if err = DropDb(ctx, pgSuperUserPgDb, dbConf.Database); err != nil {
			return fmt.Errorf("DropDb failed for database: %v: %w", dbConf.Database, err)
		}
	}

	infoLog.Info("Creating database " + dbConf.Database)
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
	infoLog.Info("Adding extensions, if any")
	if err = ExecuteFile(ctx, pgSuperUserDb, "extensions.sql", sqlAssets, replacements, infoLog); err != nil {
		return fmt.Errorf("ExecuteFile failed (extensions): %w", err)
	}

	// grant all rights on the db to the db owner user
	infoLog.Info("Granting all rights to " + dbOwnerConf.Name)
	if err = GrantAll(ctx, pgSuperUserDb, dbOwnerConf, dbConf.Database); err != nil {
		return fmt.Errorf("GrantAll failed for all rights for user: %v: %w", dbOwnerConf.Name, err)
	}

	// add other security permissions from file
	if addSecurityPermissions {
		infoLog.Info("Adding security permissions")
		if err = ExecuteFile(ctx, pgSuperUserDb, "security_permissions.sql", sqlAssets, replacements, infoLog); err != nil {
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

	infoLog.Info("Populating database")
	if err = PopulateDb(ctx, dbOwnerUserDb, sqlAssets, dbConf.SchemaCreationOrder, replacements, infoLog); err != nil {
		return fmt.Errorf("PopulateDb failed: %w", err)
	}

	infoLog.Info("Analyze")
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
	infoLog *slog.Logger) (err error) {

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
	if err = ExecuteFile(ctx, db, "schemas.sql", sqlAssets, replacements, infoLog); err != nil {
		return fmt.Errorf("ExecuteFile failed (schemas): %w", err)
	}

	// add assets (in order) which should be added before functions
	assetTypes := []string{"types", "domains", "sequences", "tables"}
	for _, assetType := range assetTypes {
		for _, schema := range schemaCreationOrder {
			assetPath := fmt.Sprintf("%s/%s_%s.sql", schema, schema, assetType)
			if err = ExecuteFile(ctx, db, assetPath, sqlAssets, replacements, infoLog); err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					continue
				}
				return fmt.Errorf("ExecuteFile failed for schema: %v, asset type: %v: %w", schema, assetType, err)
			}
		}
	}

	// add trigger funcs, then normal funcs
	funcTypes := []string{"tf_", "f_"}
	for _, funcType := range funcTypes {
		for _, schema := range schemaCreationOrder {
			dirEntries, err := sqlAssets.ReadDir("" + schema)
			if err != nil {
				return fmt.Errorf("sqlAssets.ReadDir failed for schema: %v: %w", schema, err)
			}
			for _, dirEntry := range dirEntries {
				if strings.HasPrefix(dirEntry.Name(), funcType) {
					if err = ExecuteFile(ctx, db, schema+"/"+dirEntry.Name(), sqlAssets, replacements, infoLog); err != nil {
						return fmt.Errorf("ExecuteFile failed for schema: %v, func type: %v: %w", schema, funcType, err)
					}
				}
			}
		}
	}

	// add views
	assetTypes = []string{"views", "materialized_views", "views_post_mv"}
	for _, assetType := range assetTypes {
		for _, schema := range schemaCreationOrder {

			assetPath := fmt.Sprintf("%s/%s_%s.sql", schema, schema, assetType)
			if err = ExecuteFile(ctx, db, assetPath, sqlAssets, replacements, infoLog); err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					continue
				}
				return fmt.Errorf("ExecuteFile failed for schema: %v, asset type: %v: %w", schema, assetType, err)
			}
		}
	}

	// add live then test data, if any
	assetTypes = []string{"data", "test_data"}
	for _, assetType := range assetTypes {
		for _, schema := range schemaCreationOrder {
			assetPath := fmt.Sprintf("%s/%s_%s.sql", schema, schema, assetType)
			if err = ExecuteFile(ctx, db, assetPath, sqlAssets, replacements, infoLog); err != nil {
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
					if err = ExecuteFile(ctx, db, schema+"/"+dirEntry.Name(), sqlAssets, replacements, infoLog); err != nil {
						return fmt.Errorf("ExecuteFile failed for schema: %v, asset type: %v: %w", schema, assetType, err)
					}
				}
			}
		}
	}

	return nil
}
