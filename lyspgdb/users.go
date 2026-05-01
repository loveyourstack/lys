package lyspgdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateUserIfNotExists(ctx context.Context, db *pgxpool.Pool, userConf User) (err error) {

	if userConf.Name == "" || userConf.Password == "" {
		return fmt.Errorf("user or password is empty")
	}

	// Postgres provides no IF NOT EXISTS option for CREATE USER, so attempt to create user and check for duplicate object error if it fails

	var quotedPassword string
	err = db.QueryRow(ctx, "SELECT quote_literal($1)", userConf.Password).Scan(&quotedPassword)
	if err != nil {
		return fmt.Errorf("quote_literal failed: %w", err)
	}

	stmt := fmt.Sprintf("CREATE USER %s WITH PASSWORD %s;", pgx.Identifier{userConf.Name}.Sanitize(), quotedPassword)
	_, err = db.Exec(ctx, stmt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.DuplicateObject {
			return nil
		}
		return fmt.Errorf("db.Exec failed: %w", err)
	}

	fmt.Printf("added user: %s\n", userConf.Name)

	return nil
}
