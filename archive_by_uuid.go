package lys

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// iArchiveable is a store that can be used by Archive and Restore
type iArchiveableByUuid interface {
	ArchiveByUuid(ctx context.Context, tx pgx.Tx, id uuid.UUID) (stmt string, err error)
	RestoreByUuid(ctx context.Context, tx pgx.Tx, id uuid.UUID) (stmt string, err error)
}

// Archive handles moving a record from the supplied store into its archived table
func ArchiveByUuid(env Env, db *pgxpool.Pool, store iArchiveableByUuid) http.HandlerFunc {
	return MoveRecordsByUuid(env, db, store.ArchiveByUuid, DataArchived)
}

// Restore handles moving a record from the store's archived table back to the main table
func RestoreByUuid(env Env, db *pgxpool.Pool, store iArchiveableByUuid) http.HandlerFunc {
	return MoveRecordsByUuid(env, db, store.RestoreByUuid, DataRestored)
}

// MoveRecordsById handles moving record(s) back and forth between the main table and its corresponding archived table
func MoveRecordsByUuid(env Env, db *pgxpool.Pool, moveFunc func(context.Context, pgx.Tx, uuid.UUID) (string, error), msg string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get the uuid and ensure it is valid
		uuidStr := mux.Vars(r)["id"]
		if uuidStr == "" {
			HandleUserError(http.StatusBadRequest, ErrDescIdMissing, w)
			return
		}

		idUu, err := uuid.Parse(uuidStr)
		if err != nil {
			HandleUserError(http.StatusBadRequest, ErrDescIdNotAUuid, w)
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
		stmt, err := moveFunc(r.Context(), tx, idUu)
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
				HandlePostgresError(r.Context(), stmt, "MoveRecordsByUuid: moveFunc", pgErr, env.ErrorLog, w)
				return
			}

			// unknown db error
			HandleDbError(r.Context(), stmt, fmt.Errorf("MoveRecordsByUuid: moveFunc failed: %w", err), env.ErrorLog, w)
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
