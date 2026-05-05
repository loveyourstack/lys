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
	"github.com/loveyourstack/lys/lyserr"
)

// HandleInternalError returns a generic error message to the API user and logs the error
func HandleInternalError(ctx context.Context, err error, errorLog *slog.Logger, w http.ResponseWriter) {

	resp := StdResponse{
		Status:         ReqFailed,
		ErrDescription: "An internal error occurred",
	}
	JsonResponse(resp, http.StatusInternalServerError, w)

	logError(ctx, err, errorLog)
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
	JsonResponse(resp, http.StatusBadGateway, w) // BadGateway is used to indicate that the error was caused by a 3rd party API call

	logError(ctx, err, errorLog)
}

// HandleDbError returns a specific error message to the API user if the error is caused by a bad input, e.g. a check or uniqueness violation.
// Otherwise it returns a generic error message to the API user and logs the specific error
func HandleDbError(ctx context.Context, line int, stmt string, err error, errorLog *slog.Logger, w http.ResponseWriter) {

	// if an input line number is provided, include it in all messages
	lineTxt := ""
	if line > 0 {
		lineTxt = fmt.Sprintf("line %d: ", line)
	}

	pgErrTxt := ""
	statusCode := http.StatusBadRequest

	// see if err can be unwrapped to a pgx PgError
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {

		switch pgErr.Code {

		// define text and status for errors that can be attributed to bad requests or conflicts
		case pgerrcode.CheckViolation:
			pgErrTxt = fmt.Sprintf("check constraint violation: %s", pgErr.ConstraintName)
		case pgerrcode.ExclusionViolation:
			pgErrTxt = fmt.Sprintf("exclusion constraint violation: %s", pgErr.Detail)
			statusCode = http.StatusConflict
		case pgerrcode.ForeignKeyViolation:
			pgErrTxt = fmt.Sprintf("foreign key violation: %s", pgErr.Detail)
		case pgerrcode.InvalidTextRepresentation: // e.g. enum value does not exist
			pgErrTxt = fmt.Sprintf("invalid text: %s", pgErr.Message)
		case pgerrcode.NotNullViolation:
			pgErrTxt = fmt.Sprintf("missing required field: %s", pgErr.ColumnName)
		case pgerrcode.StringDataRightTruncationDataException: // e.g. text too long for varchar(i) column
			pgErrTxt = pgErr.Message
		case pgerrcode.UndefinedObject: // e.g. enum type does not exist
			pgErrTxt = fmt.Sprintf("undefined object: %s", pgErr.Message)
		case pgerrcode.UniqueViolation:
			errStr := pgErr.ConstraintName // single column unique key
			if pgErr.Detail != "" {
				errStr = pgErr.Detail // is better, but only filled on multiple column unique keys
			}
			pgErrTxt = fmt.Sprintf("unique constraint violation: %s", errStr)
			statusCode = http.StatusConflict
		default:
			// other pgx error: treat as unknown db error
		}
	}

	// known pgx error attributable to bad input: return specific message to user, error is not logged
	if pgErrTxt != "" {
		HandleUserError(lyserr.User{Message: fmt.Sprintf("%s%s", lineTxt, pgErrTxt), StatusCode: statusCode}, w)
		return
	}

	// unknown db error
	resp := StdResponse{
		Status:         ReqFailed,
		ErrDescription: fmt.Sprintf("%sA database error occurred", lineTxt),
	}
	JsonResponse(resp, http.StatusInternalServerError, w)

	extra := []slog.Attr{slog.String("stmt", stmt)}
	if line > 0 {
		extra = append(extra, slog.Int("line", line))
	}
	logError(ctx, err, errorLog, extra...)
}

// HandleError is the general method for handling API errors where err could contain wrapped errors of other types
func HandleError(ctx context.Context, err error, errorLog *slog.Logger, w http.ResponseWriter) {

	// expected error: request canceled
	// ctx.Err() checks context state directly
	// err checks for wrapped cancellation from other errors, e.g. from db calls
	if errors.Is(ctx.Err(), context.Canceled) || errors.Is(err, context.Canceled) {
		return
	}

	// expected specific pgx errors
	if errors.Is(err, pgx.ErrNoRows) {
		HandleUserError(lyserr.User{Message: "row(s) not found", StatusCode: http.StatusNotFound}, w)
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
		HandleDbError(ctx, dbErr.Line, dbErr.Stmt, err, errorLog, w)
		return
	}

	// unknown internal error
	HandleInternalError(ctx, err, errorLog, w)
}

// logError is a helper function to log errors with user information and any extra attributes
func logError(ctx context.Context, err error, errorLog *slog.Logger, extra ...slog.Attr) {
	args := []any{slog.String("user", GetUserNameFromCtx(ctx))}
	for _, a := range extra {
		args = append(args, a)
	}
	errorLog.Error(err.Error(), args...)
}
