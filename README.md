# lys - LoveYourStack

Packages for rapid development of REST APIs handling database CRUD actions.

Only available for PostgreSQL. Most suitable for "database-first" Go developers.

## Example usage

### Define store package ([wiki](https://github.com/loveyourstack/lys/wiki/Creating-stores))

A store package contains database access functions for a specific table or view, in this case the "category" table in the "core" schema.

Boilerplate is minimized through the optional use of generic database CRUD functions.

```go
package corecategory

// define constants for this database table, which get passed to generic database functions below
const (
	schemaName     string = "core"
	tableName      string = "category"
	viewName       string = "category"
	pkColName      string = "id"
	defaultOrderBy string = "name"
)

// columns required when creating or updating a record
type Input struct {
	Name string `db:"name" json:"name,omitempty" validate:"required"`
}

// columns outputted when selecting a record. Note that Input is embedded
type Model struct {
	Id    int64 `db:"id" json:"id"`
	Input
}

type Store struct {
	Db *pgxpool.Pool
}

// define functions for this table as methods of the Store struct
// use lyspg generic functions if possible, but can also write custom implementations

func (s Store) Delete(ctx context.Context, id int64) error {
	return lyspg.DeleteUnique(ctx, s.Db, schemaName, tableName, pkColName, id)
}

func (s Store) Insert(ctx context.Context, input Input) (newId int64, err error) {
	return lyspg.Insert[Input, int64](ctx, s.Db, schemaName, tableName, pkColName, input)
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, gDbTags, params)
}

// etc

```

### Create routes ([wiki](https://github.com/loveyourstack/lys/wiki/Creating-routes))

Pass the store package to generic GET, POST, etc handlers to get full REST API CRUD functionality for this table with minimal boilerplate.

```go
package main

func (srvApp *httpServerApplication) getRoutes(apiEnv lys.Env) http.Handler {

	endpoint := "/core-categories"

	// get full CRUD functionality using lys generic handlers, passing the store defined above
	// no framework: free to write custom handlers when needed

	categoryStore := corecategory.Store{Db: srvApp.Db}
	r.HandleFunc(endpoint, lys.Get[corecategory.Model](apiEnv, categoryStore)).Methods("GET")
	r.HandleFunc(endpoint+"/{id}", lys.GetById[corecategory.Model](apiEnv, categoryStore)).Methods("GET")
	r.HandleFunc(endpoint, lys.Post[corecategory.Input, int64](apiEnv, categoryStore)).Methods("POST")
	r.HandleFunc(endpoint+"/{id}", lys.Put[corecategory.Input](apiEnv, categoryStore)).Methods("PUT")
	r.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, categoryStore)).Methods("PATCH")
	r.HandleFunc(endpoint+"/{id}", lys.Delete(apiEnv, categoryStore)).Methods("DELETE")
}

```

### Use routes

We can now start the HTTP server app and use the routes above.

```console
curl localhost:8010/core-categories?name=Seafood
curl localhost:8010/core-categories/1
curl --header "Content-Type: application/json" --request POST --data '{"name":"Fruit"}' localhost:8010/core-categories
# etc
```

See the [Northwind sample application](https://github.com/loveyourstack/northwind) for a complete application using these packages.

## Features

* Library only: is not a framework, and does not use code generation, so can be overriden at every step to deal with exceptional cases
* Support for GET many, GET single, POST, PUT, PATCH and DELETE
* Support for [sorting, paging and filtering GET results](https://github.com/loveyourstack/lys/wiki/GET-request-URL-parameters) via customizable URL params
* Uses [pgx](https://github.com/jackc/pgx/) for database access and only uses parameterized SQL queries
* Support for Excel and CSV output
* Uses generics and reflection to minimize boilerplate
* Custom date/time types with zero default values and sensible JSON formats
* Fast rowcount function, including estimated count for large tables with query conditions
* Struct validation using [validator](https://github.com/go-playground/validator)
* Distinction between user errors (unlogged, reported to user) and application errors (logged, hidden from user)
* Provides useful bulk insert (COPY) wrapper, and bulk update/delete (batch) wrappers
* Support for getting and filtering enum values
* Support for selection from database set-returning functions
* Database creation function from embedded SQL files
* Archive (soft delete) + restore functions
* and more. See the [wiki](https://github.com/loveyourstack/lys/wiki)

## Current limitations

* Only supports PostgreSQL
* No database obfuscation. Struct "db" tags must be added and must be identical to the "json" tag, unless the latter is "-"
* Limited support for database date/time arrays

## Testing

See CONTRIBUTING.md for setup instructions.

## Supported Go and PostgreSQL Versions

Preliminary values:

Go 1.16+ (due to embed.FS)

PostgreSQL 13+ (due to gen_random_uuid)