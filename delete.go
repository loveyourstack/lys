package lys

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// iDeletable is a store that can be used by Delete
type iDeletable interface {
	Delete(ctx context.Context, id int64) (stmt string, err error)
}

// Delete handles deletion of a single item using the supplied store
func Delete(env Env, store iDeletable) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get the Id and ensure it is an int
		vars := mux.Vars(r)
		id, err := strconv.ParseInt(vars["id"], 10, 64)
		if err != nil {
			HandleUserError(http.StatusBadRequest, ErrDescIdNotAnInteger, w)
			return
		}

		// delete item from db
		stmt, err := store.Delete(r.Context(), id)
		if err != nil {

			// expected error: request canceled
			if errors.Is(err, context.Canceled) {
				return
			}

			// expected error: Id does not exist
			if errors.Is(err, pgx.ErrNoRows) {
				HandleUserError(http.StatusBadRequest, ErrDescInvalidId, w)
				return
			}

			// expected error: Id not unique (shouldn't happen)
			if errors.Is(err, pgx.ErrTooManyRows) {
				HandleUserError(http.StatusBadRequest, ErrDescIdNotUnique, w)
				return
			}

			// handle errors from postgres
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				HandlePostgresError(r.Context(), stmt, "Delete: store.Delete", pgErr, env.ErrorLog, w)
				return
			}

			// unknown db error
			HandleDbError(r.Context(), stmt, fmt.Errorf("Delete: store.Delete failed: %w", err), env.ErrorLog, w)
			return
		}

		// success
		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   DataDeleted,
		}
		JsonResponse(resp, http.StatusOK, nil, w)
	}
}
