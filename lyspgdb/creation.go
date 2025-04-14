package lyspgdb

import (
	"context"
	"embed"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateLocalDb creates or recreates a test or dev db
func CreateLocalDb(ctx context.Context, sqlAssets embed.FS, dbConf Database, dbSuperUser, dbOwnerConf User,
	dropExisting, addSecurityPermissions bool, replacementsMap map[string]string, infoLog *slog.Logger) (err error) {

	pgDbConf := Database{
		Host:     dbConf.Host,
		Port:     dbConf.Port,
		Database: "postgres",
	}

	pgUserConf := User{
		Name:     dbSuperUser.Name,
		Password: dbSuperUser.Password,
	}

	// connect with superuser to postgres db
	pgUserPgDb, err := GetPool(ctx, pgDbConf, pgUserConf, "test")
	if err != nil {
		return fmt.Errorf("GetPool failed (postgres db with %v user): %w", dbSuperUser.Name, err)
	}

	// (re-)create database
	if dropExisting {
		infoLog.Info("Dropping database " + dbConf.Database + " if exists")
		if err = DropDb(ctx, pgUserPgDb, dbConf.Database); err != nil {
			return fmt.Errorf("DropDb failed for database: %v: %w", dbConf.Database, err)
		}
	}

	infoLog.Info("Creating database " + dbConf.Database)
	if err = CreateDb(ctx, pgUserPgDb, dbConf.Database); err != nil {
		return fmt.Errorf("CreateDb failed for database: %v: %w", dbConf.Database, err)
	}
	pgUserPgDb.Close()

	// ----------------

	// connect with superuser user to target db
	pgUserDb, err := GetPool(ctx, dbConf, pgUserConf, "CreateLocalDb func")
	if err != nil {
		return fmt.Errorf("GetPool failed (database %v with %v user): %w", dbConf.Database, dbSuperUser.Name, err)
	}

	// add database extensions
	infoLog.Info("Adding extensions, if any")
	if err = ExecuteFile(ctx, pgUserDb, "extensions.sql", sqlAssets, replacementsMap, infoLog); err != nil {
		return fmt.Errorf("ExecuteFile failed (extensions): %w", err)
	}

	// grant all rights on the db to the db owner user
	infoLog.Info("Granting all rights to " + dbOwnerConf.Name)
	if err = GrantAll(ctx, pgUserDb, dbOwnerConf, dbConf.Database); err != nil {
		return fmt.Errorf("GrantAll failed for all rights for user: %v: %w", dbOwnerConf.Name, err)
	}

	pgUserDb.Close()

	// ----------------

	// connect with db owner user to target db
	dbOwnerUserDb, err := GetPool(ctx, dbConf, dbOwnerConf, "CreateLocalDb func")
	if err != nil {
		return fmt.Errorf("GetPool failed (database %v with %v user): %w", dbConf.Database, dbOwnerConf.Name, err)
	}
	defer dbOwnerUserDb.Close()

	// ----------------

	// add other security permissions from file
	if addSecurityPermissions {
		infoLog.Info("Adding security permissions")
		if err = ExecuteFile(ctx, pgUserDb, "security_permissions.sql", sqlAssets, replacementsMap, infoLog); err != nil {
			return fmt.Errorf("ExecuteFile failed (security_permissions): %w", err)
		}
	}

	// ----------------

	// populate and analyze db

	infoLog.Info("Populating database")
	if err = PopulateDb(ctx, dbOwnerUserDb, sqlAssets, dbConf.SchemaCreationOrder, replacementsMap, infoLog); err != nil {
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
	qry := "DROP DATABASE IF EXISTS " + dbName + " WITH (FORCE);"
	if _, err = pgUserPgDb.Exec(ctx, qry); err != nil {
		return fmt.Errorf("pgUserPgDb.Exec failed: %w", err)
	}

	return nil
}

// CreateDb creates a database
// pgUserPgDb is a connection to the postgres database with the postgres user
func CreateDb(ctx context.Context, pgUserPgDb *pgxpool.Pool, dbName string) (err error) {

	// create the database
	qry := "CREATE DATABASE " + dbName + ";"
	if _, err = pgUserPgDb.Exec(ctx, qry); err != nil {
		return fmt.Errorf("pgUserPgDb.Exec failed: %w", err)
	}

	return nil
}

// GrantAll grants all rights to the specified user on the specified db
// pgUserDb is a connection to the target database with the postgres user
func GrantAll(ctx context.Context, pgUserDb *pgxpool.Pool, userConf User, dbName string) (err error) {

	// grant all rights on this database
	qry := "GRANT ALL ON DATABASE " + dbName + " TO " + userConf.Name + ";"
	if _, err = pgUserDb.Exec(ctx, qry); err != nil {
		return fmt.Errorf("pgUserDb.Exec (grant all on database) failed: %w", err)
	}

	// grant access to public schema
	qry = "GRANT ALL ON SCHEMA public TO " + userConf.Name + ";"
	if _, err = pgUserDb.Exec(ctx, qry); err != nil {
		return fmt.Errorf("pgUserDb.Exec (grant all on schema public) failed: %w", err)
	}

	return nil
}

// PopulateDb writes schema, tables, functions and views
func PopulateDb(ctx context.Context, db *pgxpool.Pool, sqlAssets embed.FS, schemaCreationOrder []string, replacementsMap map[string]string,
	infoLog *slog.Logger) (err error) {

	// make sure db is empty (no non-system tables)
	qry := `SELECT count(*) FROM pg_class c
	  JOIN pg_namespace s ON s.oid = c.relnamespace
	  WHERE s.nspname NOT IN ('pg_catalog', 'information_schema')
	  AND s.nspname NOT LIKE 'pg_temp%' AND c.relname NOT LIKE 'pg_%'`

	var count int
	row := db.QueryRow(ctx, qry)
	if err = row.Scan(&count); err != nil {
		return fmt.Errorf("row.Scan failed: %w", err)
	}
	if count != 0 {
		return fmt.Errorf("target database is not empty, it contains %d tables", count)
	}

	// create schemas and assign default rights
	if err = ExecuteFile(ctx, db, "schemas.sql", sqlAssets, replacementsMap, infoLog); err != nil {
		return fmt.Errorf("ExecuteFile failed (schemas): %w", err)
	}

	// add assets (in order) which should be added before functions
	assetTypes := []string{"types", "domains", "sequences", "tables"}
	for _, assetType := range assetTypes {
		for _, schema := range schemaCreationOrder {
			_, err := sqlAssets.ReadFile("" + schema + "/" + schema + "_" + assetType + ".sql")
			if err != nil {
				// file not found, skip
				continue
			}
			if err = ExecuteFile(ctx, db, schema+"/"+schema+"_"+assetType+".sql", sqlAssets, replacementsMap, infoLog); err != nil {
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
				if dirEntry.Name()[:len(funcType)] == funcType {
					if err = ExecuteFile(ctx, db, schema+"/"+dirEntry.Name(), sqlAssets, replacementsMap, infoLog); err != nil {
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
			_, err := sqlAssets.ReadFile("" + schema + "/" + schema + "_" + assetType + ".sql")
			if err != nil {
				continue
			}
			if err = ExecuteFile(ctx, db, schema+"/"+schema+"_"+assetType+".sql", sqlAssets, replacementsMap, infoLog); err != nil {
				return fmt.Errorf("ExecuteFile failed for schema: %v, asset type: %v: %w", schema, assetType, err)
			}
		}
	}

	// add live then test data, if any
	assetTypes = []string{"data", "test_data"}
	for _, assetType := range assetTypes {
		for _, schema := range schemaCreationOrder {
			_, err := sqlAssets.ReadFile("" + schema + "/" + schema + "_" + assetType + ".sql")
			if err != nil {
				continue
			}
			if err = ExecuteFile(ctx, db, schema+"/"+schema+"_"+assetType+".sql", sqlAssets, replacementsMap, infoLog); err != nil {
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
				if dirEntry.Name()[:len(assetType)] == assetType {
					if err = ExecuteFile(ctx, db, schema+"/"+dirEntry.Name(), sqlAssets, replacementsMap, infoLog); err != nil {
						return fmt.Errorf("ExecuteFile failed for schema: %v, asset type: %v: %w", schema, assetType, err)
					}
				}
			}
		}
	}

	return nil
}
