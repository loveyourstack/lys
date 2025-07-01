package lyspgdb

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateUserIfNotExists(ctx context.Context, db *pgxpool.Pool, userConf User, options []string) (err error) {

	if userConf.Name == "" || userConf.Password == "" {
		return fmt.Errorf("user or password is empty")
	}

	stmt := fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s' %s;", userConf.Name, userConf.Password, strings.Join(options, " "))
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
