package lys

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// iArchiveableByUuid is a store that can be used by ArchiveByUuid and RestoreByUuid
type iArchiveableByUuid interface {
	ArchiveByUuid(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
	RestoreByUuid(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
}

// ArchiveByUuid handles moving a record from the supplied store into its archived table
func ArchiveByUuid(env Env, db *pgxpool.Pool, store iArchiveableByUuid) http.HandlerFunc {
	return MoveRecordsByUuid(env, db, store.ArchiveByUuid, DataArchived)
}

// RestoreByUuid handles moving a record from the store's archived table back to the main table
func RestoreByUuid(env Env, db *pgxpool.Pool, store iArchiveableByUuid) http.HandlerFunc {
	return MoveRecordsByUuid(env, db, store.RestoreByUuid, DataRestored)
}

// MoveRecordsByUuid handles moving record(s) back and forth between the main table and its corresponding archived table
func MoveRecordsByUuid(env Env, db *pgxpool.Pool, moveFunc func(context.Context, pgx.Tx, uuid.UUID) error, msg string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get the uuid and ensure it is valid
		uuidStr := mux.Vars(r)["id"]
		if uuidStr == "" {
			HandleUserError(ErrIdMissing, w)
			return
		}

		idUu, err := uuid.Parse(uuidStr)
		if err != nil {
			HandleUserError(ErrIdNotAUuid, w)
			return
		}

		// begin tx
		tx, err := db.Begin(r.Context())
		if err != nil {
			HandleInternalError(r.Context(), fmt.Errorf("MoveRecordsByUuid: db.Begin failed: %w", err), env.ErrorLog, w)
			return
		}
		defer tx.Rollback(r.Context())

		// try the operation
		err = moveFunc(r.Context(), tx, idUu)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("MoveRecordsByUuid: moveFunc failed: %w", err), env.ErrorLog, w)
			return
		}

		// success: commit tx
		err = tx.Commit(r.Context())
		if err != nil {
			HandleInternalError(r.Context(), fmt.Errorf("MoveRecordsByUuid: tx.Commit failed: %w", err), env.ErrorLog, w)
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
