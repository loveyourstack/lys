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
	"github.com/jackc/pgx/v5/pgxpool"
)

// iArchiveable is a store that can be used by Archive and Restore
type iArchiveable interface {
	Archive(ctx context.Context, tx pgx.Tx, id int64) (stmt string, err error)
	Restore(ctx context.Context, tx pgx.Tx, id int64) (stmt string, err error)
}

// Archive handles moving a record from the supplied store into its archived table
func Archive(env Env, db *pgxpool.Pool, store iArchiveable) http.HandlerFunc {
	return MoveRecordsById(env, db, store.Archive, DataArchived)
}

// Restore handles moving a record from the store's archived table back to the main table
func Restore(env Env, db *pgxpool.Pool, store iArchiveable) http.HandlerFunc {
	return MoveRecordsById(env, db, store.Restore, DataRestored)
}

// MoveRecordsById handles moving record(s) back and forth between the main table and its corresponding archived table
func MoveRecordsById(env Env, db *pgxpool.Pool, moveFunc func(context.Context, pgx.Tx, int64) (string, error), msg string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get the id and ensure it is an int
		vars := mux.Vars(r)
		id, err := strconv.ParseInt(vars["id"], 10, 64)
		if err != nil {
			HandleUserError(http.StatusBadRequest, ErrDescIdNotAnInteger, w)
			return
		}

		// begin tx
		tx, err := db.Begin(r.Context())
		if err != nil {
			HandleInternalError(r.Context(), fmt.Errorf("MoveRecordsById: db.Begin failed: %w", err), env.ErrorLog, w)
			return
		}
		defer tx.Rollback(r.Context())

		// try the operation
		stmt, err := moveFunc(r.Context(), tx, id)
		if err != nil {

			// expected error: request canceled
			if errors.Is(err, context.Canceled) {
				return
			}

			// expected error: id does not exist
			if errors.Is(err, pgx.ErrNoRows) {
				HandleUserError(http.StatusBadRequest, ErrDescInvalidId, w)
				return
			}

			// handle errors from postgres
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				HandlePostgresError(r.Context(), stmt, "MoveRecordsById: moveFunc", pgErr, env.ErrorLog, w)
				return
			}

			// unknown db error
			HandleDbError(r.Context(), stmt, fmt.Errorf("MoveRecordsById: moveFunc failed: %w", err), env.ErrorLog, w)
			return
		}

		// success: commit tx
		err = tx.Commit(r.Context())
		if err != nil {
			HandleInternalError(r.Context(), fmt.Errorf("MoveRecordsById: tx.Commit failed: %w", err), env.ErrorLog, w)
			return
		}

		// success
		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   msg,
		}
		JsonResponse(resp, http.StatusOK, w)
	}
}
