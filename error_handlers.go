package lys

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// HandleDbError returns a generic error message to the API user and includes the failing statement in the error log
func HandleDbError(ctx context.Context, stmt string, err error, errorLog *slog.Logger, w http.ResponseWriter) {

	resp := StdResponse{
		Status:         ReqFailed,
		ErrDescription: "A database error occurred",
	}
	JsonResponse(resp, http.StatusInternalServerError, nil, w)

	userName := GetUserNameFromCtx(ctx, "Unknown")
	errorLog.Error(err.Error(), slog.String("user", userName), slog.String("stmt", stmt))
}

// HandleInternalError returns a generic error message to the API user and logs the error
func HandleInternalError(ctx context.Context, err error, errorLog *slog.Logger, w http.ResponseWriter) {

	resp := StdResponse{
		Status:         ReqFailed,
		ErrDescription: "An internal error occurred",
	}
	JsonResponse(resp, http.StatusInternalServerError, nil, w)

	userName := GetUserNameFromCtx(ctx, "Unknown")
	errorLog.Error(err.Error(), slog.String("user", userName))
}

// HandlePostgresError handles an error that is of type PgError
func HandlePostgresError(ctx context.Context, stmt, callerFunc string, pgErr *pgconn.PgError, errorLog *slog.Logger, w http.ResponseWriter) {

	switch pgErr.Code {

	// expected errors
	case pgerrcode.CheckViolation:
		HandleUserError(http.StatusBadRequest, "check constraint violation: "+pgErr.ConstraintName, w)
		return
	case pgerrcode.ForeignKeyViolation:
		HandleUserError(http.StatusBadRequest, "foreign key violation: "+pgErr.Detail, w)
		return
	case pgerrcode.InvalidTextRepresentation: // e.g. enum value does not exist
		HandleUserError(http.StatusBadRequest, "invalid text: "+pgErr.Message, w)
		return
	case pgerrcode.UndefinedObject: // e.g. enum type does not exist
		HandleUserError(http.StatusBadRequest, "undefined object: "+pgErr.Message, w)
		return
	case pgerrcode.UniqueViolation:
		HandleUserError(http.StatusConflict, "unique constraint violation: "+pgErr.Detail, w)
		return

	// unexpected postgres db error or db process cancelation (pgerrcode.QueryCanceled)
	default:
		HandleDbError(ctx, stmt, fmt.Errorf(callerFunc+" failed: %w", pgErr), errorLog, w)
		return
	}
}

// HandleUserError returns a helpful message to the API user, but does not log the error
func HandleUserError(statusCode int, userErrMsg string, w http.ResponseWriter) {

	resp := StdResponse{
		Status:         ReqFailed,
		ErrDescription: userErrMsg,
	}
	JsonResponse(resp, statusCode, nil, w)
}
