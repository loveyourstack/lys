package lyspg

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/loveyourstack/lys/lyserr"
)

// Sleep creates an artifical longrunning query in the db which can be viewed using pg_stat_activity.
// It is used for testing context and database cancelation
func Sleep(ctx context.Context, db PoolOrTx, secs int) (err error) {

	stmt := fmt.Sprintf("SELECT pg_sleep(%d);", secs)

	rows, _ := db.Query(ctx, stmt)
	_, err = pgx.CollectExactlyOneRow(rows, pgx.RowTo[string])
	if err != nil {

		// request canceled via context. Test by starting a request then canceling it (e.g. using Postman)
		if errors.Is(ctx.Err(), context.Canceled) || errors.Is(err, context.Canceled) {
			fmt.Println("canceled via context")
			return err
		}

		// deadline exceeded via context. Test by setting a context timeout
		if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("canceled due to context deadline exceeded")
			return err
		}

		// db pid canceled. Test by using pg_cancel_backend(pid)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.QueryCanceled {
			fmt.Println("canceled in database")
			return err
		}

		// unknown db error
		return lyserr.Db{Err: fmt.Errorf("pgx.CollectExactlyOneRow failed: %w", err), Stmt: stmt}
	}

	fmt.Printf("slept %d seconds\n", secs)
	return nil
}
