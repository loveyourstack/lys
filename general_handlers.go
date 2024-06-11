package lys

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/loveyourstack/lys/lyspg"
)

// Message returns the supplied msg in the Data field
func Message(msg string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   msg,
		}
		JsonResponse(resp, http.StatusOK, w)
	}
}

// NotFound provides a response informing the user that the requested route was not found
func NotFound() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		resp := StdResponse{
			Status:         ReqFailed,
			ErrDescription: ErrDescRouteNotFound,
		}
		JsonResponse(resp, http.StatusNotFound, w)
	}
}

// PgSleep creates an artifical longrunning query in the db which can be viewed using pg_stat_activity
// used for testing context cancelation
func PgSleep(db lyspg.PoolOrTx, errorLog *slog.Logger, secs int) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		stmt := fmt.Sprintf("SELECT pg_sleep(%d);", secs)

		rows, _ := db.Query(r.Context(), stmt)
		_, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[string])
		if err != nil {

			// request canceled: cancelation propagated to db via context
			if errors.Is(err, context.Canceled) {
				return
			}

			// db pid canceled (pg_cancel_backend(pid))
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.QueryCanceled {
				resp := StdResponse{
					Status:         ReqFailed,
					ErrDescription: "process canceled in database",
				}
				JsonResponse(resp, http.StatusInternalServerError, w)
				return
			}

			// unknown db error
			HandleDbError(r.Context(), stmt, fmt.Errorf("PgSleep: pgx.CollectExactlyOneRow failed: %w", err), errorLog, w)
			return
		}

		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   fmt.Sprintf("slept %d seconds", secs),
		}
		JsonResponse(resp, http.StatusOK, w)
	}
}
