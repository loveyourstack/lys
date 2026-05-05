package lys

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lyspg"
)

// iArchiveableById is a store that can be used by ArchiveById and RestoreById
type iArchiveableById interface {
	ArchiveById(ctx context.Context, tx pgx.Tx, id int64) error
	RestoreById(ctx context.Context, tx pgx.Tx, id int64) error
}

// iArchiveableByUuid is a store that can be used by ArchiveByUuid and RestoreByUuid
type iArchiveableByUuid interface {
	ArchiveByUuid(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
	RestoreByUuid(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
}

var archiveParseIdFunc = func(idStr string) (int64, error) {
	return strconv.ParseInt(idStr, 10, 64)
}

var archiveParseUuidFunc = func(idStr string) (uuid.UUID, error) {
	return uuid.Parse(idStr)
}

// ArchiveById handles moving a record from the supplied store into its archived table
func ArchiveById(env Env, db *pgxpool.Pool, store iArchiveableById) http.HandlerFunc {
	return moveRecords(env, db, store.ArchiveById, archiveParseIdFunc, "ArchiveById", DataArchived)
}

// ArchiveByUuid handles moving a record from the supplied store into its archived table
func ArchiveByUuid(env Env, db *pgxpool.Pool, store iArchiveableByUuid) http.HandlerFunc {
	return moveRecords(env, db, store.ArchiveByUuid, archiveParseUuidFunc, "ArchiveByUuid", DataArchived)
}

// moveRecords handles moving record(s) back and forth between the main table and its corresponding archived table
func moveRecords[T lyspg.PrimaryKeyType](env Env, db *pgxpool.Pool, moveFunc func(context.Context, pgx.Tx, T) error,
	parseIdFunc func(string) (T, error), callingFunc, msg string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get the id param
		idStr := mux.Vars(r)["id"]
		if idStr == "" {
			HandleUserError(ErrIdMissing, w)
			return
		}

		// parse the id param into a T
		id, err := parseIdFunc(idStr)
		if err != nil {
			HandleUserError(ErrIdParseError, w)
			return
		}

		// begin tx
		tx, err := db.Begin(r.Context())
		if err != nil {
			HandleInternalError(r.Context(), fmt.Errorf("%s: db.Begin failed: %w", callingFunc, err), env.ErrorLog, w)
			return
		}
		defer tx.Rollback(r.Context())

		// try the operation
		err = moveFunc(r.Context(), tx, id)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("%s: moveFunc failed: %w", callingFunc, err), env.ErrorLog, w)
			return
		}

		// success: commit tx
		err = tx.Commit(r.Context())
		if err != nil {
			HandleInternalError(r.Context(), fmt.Errorf("%s: tx.Commit failed: %w", callingFunc, err), env.ErrorLog, w)
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

// RestoreById handles moving a record from the store's archived table back to the main table
func RestoreById(env Env, db *pgxpool.Pool, store iArchiveableById) http.HandlerFunc {
	return moveRecords(env, db, store.RestoreById, archiveParseIdFunc, "RestoreById", DataRestored)
}

// RestoreByUuid handles moving a record from the store's archived table back to the main table
func RestoreByUuid(env Env, db *pgxpool.Pool, store iArchiveableByUuid) http.HandlerFunc {
	return moveRecords(env, db, store.RestoreByUuid, archiveParseUuidFunc, "RestoreByUuid", DataRestored)
}
