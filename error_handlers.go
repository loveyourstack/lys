package lys

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/loveyourstack/lys/lyserr"
)

// HandleInternalError returns a generic error message to the API user and logs the error
func HandleInternalError(ctx context.Context, err error, errorLog *slog.Logger, w http.ResponseWriter) {

	resp := StdResponse{
		Status:         ReqFailed,
		ErrDescription: "An internal error occurred",
	}
	JsonResponse(resp, http.StatusInternalServerError, w)

	userName := GetUserNameFromCtx(ctx, "Unknown")
	errorLog.Error(err.Error(), slog.String("user", userName))
}

// HandleUserError returns a helpful message to the API user, but does not log the error. If HTTP status is not provided, BadRequest is assumed.
func HandleUserError(err lyserr.User, w http.ResponseWriter) {

	if err.StatusCode == 0 {
		err.StatusCode = http.StatusBadRequest
	}

	resp := StdResponse{
		Status:         ReqFailed,
		ErrDescription: err.Message,
	}
	JsonResponse(resp, err.StatusCode, w)
}

// HandleExtError returns the external message to the API user and logs the error
func HandleExtError(ctx context.Context, extMessage string, err error, errorLog *slog.Logger, w http.ResponseWriter) {

	resp := StdResponse{
		Status:         ReqFailed,
		ErrDescription: extMessage,
	}
	JsonResponse(resp, http.StatusInternalServerError, w)

	userName := GetUserNameFromCtx(ctx, "Unknown")
	errorLog.Error(err.Error(), slog.String("user", userName))
}

// HandleDbError returns a generic error message to the API user and includes the failing statement in the error log
func HandleDbError(ctx context.Context, stmt string, err error, errorLog *slog.Logger, w http.ResponseWriter) {

	// see if err can be unwrapped to a pgx PgError
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {

		switch pgErr.Code {

		// handle expected errors that can be attributed to bad requests or conflicts
		case pgerrcode.CheckViolation:
			HandleUserError(lyserr.User{Message: "check constraint violation: " + pgErr.ConstraintName}, w)
			return
		case pgerrcode.ExclusionViolation:
			HandleUserError(lyserr.User{Message: "exclusion constraint violation: " + pgErr.Detail, StatusCode: http.StatusConflict}, w)
			return
		case pgerrcode.ForeignKeyViolation:
			HandleUserError(lyserr.User{Message: "foreign key violation: " + pgErr.Detail}, w)
			return
		case pgerrcode.InvalidTextRepresentation: // e.g. enum value does not exist
			HandleUserError(lyserr.User{Message: "invalid text: " + pgErr.Message}, w)
			return
		case pgerrcode.StringDataRightTruncationDataException: // e.g. text too long for varchar(i) column
			HandleUserError(lyserr.User{Message: pgErr.Message}, w)
			return
		case pgerrcode.UndefinedObject: // e.g. enum type does not exist
			HandleUserError(lyserr.User{Message: "undefined object: " + pgErr.Message}, w)
			return
		case pgerrcode.UniqueViolation:
			HandleUserError(lyserr.User{Message: "unique constraint violation: " + pgErr.Detail, StatusCode: http.StatusConflict}, w)
			return
		}
	}

	// unknown db error
	resp := StdResponse{
		Status:         ReqFailed,
		ErrDescription: "A database error occurred",
	}
	JsonResponse(resp, http.StatusInternalServerError, w)

	userName := GetUserNameFromCtx(ctx, "Unknown")
	errorLog.Error(err.Error(), slog.String("user", userName), slog.String("stmt", stmt))
}

// HandleError is the general method for handling API errors where err could contain wrapped errors of other types
func HandleError(ctx context.Context, err error, errorLog *slog.Logger, w http.ResponseWriter) {

	// expected error: request canceled
	if errors.Is(err, context.Canceled) {
		return
	}

	// expected specific pgx errors
	if errors.Is(err, pgx.ErrNoRows) {
		HandleUserError(lyserr.User{Message: "row(s) not found"}, w)
		return
	}
	if errors.Is(err, pgx.ErrTooManyRows) {
		HandleUserError(lyserr.User{Message: "too many rows found"}, w)
		return
	}

	// see if err can be unwrapped to a userErr
	userErr := lyserr.User{}
	if errors.As(err, &userErr) {
		HandleUserError(userErr, w)
		return
	}

	// see if err can be unwrapped to an extErr
	extErr := lyserr.Ext{}
	if errors.As(err, &extErr) {
		// pass err, not extErr, to keep full trace
		HandleExtError(ctx, extErr.Message, err, errorLog, w)
		return
	}

	// see if err can be unwrapped to a dbErr or pgx PgError
	dbErr := lyserr.Db{}
	if errors.As(err, &dbErr) {
		// pass err, not dbErr, to keep full trace
		HandleDbError(ctx, dbErr.Stmt, err, errorLog, w)
		return
	}

	// unknown internal error
	HandleInternalError(ctx, err, errorLog, w)
}
