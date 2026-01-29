package lyserr

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// User is an error that is caused by the user and should be reported back to him and not logged
type User struct {
	Message    string // shown to user to help him identify error
	StatusCode int    // optional, for HTTP server: if not supplied, 400 - BadRequest will be used
}

func (e User) Error() string {
	return e.Message
}

// ---------------

// Db is an error that comes from the Postgres database. It might contain a wrapped pgx PgError.
type Db struct {
	Err  error
	Line int    // optional: the line number in input that caused the error
	Stmt string // the SQL statement that caused the error
}

func (e Db) Error() string {
	return e.Err.Error()
}

func (e Db) Unwrap() error {
	var pgErr *pgconn.PgError
	if errors.As(e.Err, &pgErr) {
		return pgErr
	}

	return e.Err
}

// ---------------

// Ext is an error that comes from an external API. It should be both logged, and the message shown to users
type Ext struct {
	Err     error
	Message string // user-readable message from the API
}

func (e Ext) Error() string {
	return e.Err.Error()
}
