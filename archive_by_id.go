package lys

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// iArchiveableById is a store that can be used by ArchiveById and RestoreById
type iArchiveableById interface {
	ArchiveById(ctx context.Context, tx pgx.Tx, id int64) error
	RestoreById(ctx context.Context, tx pgx.Tx, id int64) error
}

// ArchiveById handles moving a record from the supplied store into its archived table
func ArchiveById(env Env, db *pgxpool.Pool, store iArchiveableById) http.HandlerFunc {
	return MoveRecordsById(env, db, store.ArchiveById, DataArchived)
}

// RestoreById handles moving a record from the store's archived table back to the main table
func RestoreById(env Env, db *pgxpool.Pool, store iArchiveableById) http.HandlerFunc {
	return MoveRecordsById(env, db, store.RestoreById, DataRestored)
}

// MoveRecordsById handles moving record(s) back and forth between the main table and its corresponding archived table
func MoveRecordsById(env Env, db *pgxpool.Pool, moveFunc func(context.Context, pgx.Tx, int64) error, msg string) http.HandlerFunc {

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
		err = moveFunc(r.Context(), tx, id)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("MoveRecordsById: moveFunc failed: %w", err), env.ErrorLog, w)
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
