package lys

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lyspg"
)

type iArchiveable[idT lyspg.PrimaryKeyType] interface {
	Archive(ctx context.Context, tx pgx.Tx, id idT) error
	Restore(ctx context.Context, tx pgx.Tx, id idT) error
}

// Archive handles moving a record from the supplied store into its archived table
func Archive[idT lyspg.PrimaryKeyType](env Env, db *pgxpool.Pool, store iArchiveable[idT]) http.HandlerFunc {
	return moveRecords(env, db, store.Archive, "Archive", DataArchived)
}

// Restore handles moving a record from the store's archived table back to the main table
func Restore[idT lyspg.PrimaryKeyType](env Env, db *pgxpool.Pool, store iArchiveable[idT]) http.HandlerFunc {
	return moveRecords(env, db, store.Restore, "Restore", DataRestored)
}

// moveRecords handles moving record(s) back and forth between the main table and its corresponding archived table
func moveRecords[idT lyspg.PrimaryKeyType](env Env, db *pgxpool.Pool, moveFunc func(context.Context, pgx.Tx, idT) error,
	callingFunc, msg string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// get the id param
		idStr := mux.Vars(r)["id"]
		if idStr == "" {
			HandleUserError(ErrIdMissing, w)
			return
		}

		// parse the id param into an idT
		id, err := parseIdByType[idT](idStr)
		if err != nil {
			HandleUserError(ErrIdParseError, w)
			return
		}

		// begin tx
		tx, err := db.Begin(ctx)
		if err != nil {
			HandleInternalError(ctx, fmt.Errorf("%s: db.Begin failed: %w", callingFunc, err), env.ErrorLog, w)
			return
		}
		defer tx.Rollback(ctx)

		// try the operation
		err = moveFunc(ctx, tx, id)
		if err != nil {
			HandleError(ctx, fmt.Errorf("%s: moveFunc failed: %w", callingFunc, err), env.ErrorLog, w)
			return
		}

		// success: commit tx
		err = tx.Commit(ctx)
		if err != nil {
			HandleInternalError(ctx, fmt.Errorf("%s: tx.Commit failed: %w", callingFunc, err), env.ErrorLog, w)
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
