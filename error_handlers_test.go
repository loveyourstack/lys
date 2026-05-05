package lys

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var discardLog = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 10}))

func decodeStdResponse(t *testing.T, w *httptest.ResponseRecorder) StdResponse {
	t.Helper()
	var resp StdResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	return resp
}

// ------------------------------------------------------------------------------------------------------------------------
// HandleInternalError

func TestHandleInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	HandleInternalError(context.Background(), errors.New("something broke"), discardLog, w)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := decodeStdResponse(t, w)
	assert.Equal(t, ReqFailed, resp.Status)
	assert.Equal(t, "An internal error occurred", resp.ErrDescription)
}

// ------------------------------------------------------------------------------------------------------------------------
// HandleUserError

func TestHandleUserError_DefaultStatus(t *testing.T) {
	w := httptest.NewRecorder()
	HandleUserError(lyserr.User{Message: "bad input"}, w)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := decodeStdResponse(t, w)
	assert.Equal(t, ReqFailed, resp.Status)
	assert.Equal(t, "bad input", resp.ErrDescription)
}

func TestHandleUserError_CustomStatus(t *testing.T) {
	w := httptest.NewRecorder()
	HandleUserError(lyserr.User{Message: "permission denied", StatusCode: http.StatusForbidden}, w)

	assert.Equal(t, http.StatusForbidden, w.Code)
	resp := decodeStdResponse(t, w)
	assert.Equal(t, "permission denied", resp.ErrDescription)
}

// ------------------------------------------------------------------------------------------------------------------------
// HandleExtError

func TestHandleExtError(t *testing.T) {
	w := httptest.NewRecorder()
	HandleExtError(context.Background(), "upstream service unavailable", errors.New("dial timeout"), discardLog, w)

	assert.Equal(t, http.StatusBadGateway, w.Code)
	resp := decodeStdResponse(t, w)
	assert.Equal(t, ReqFailed, resp.Status)
	assert.Equal(t, "upstream service unavailable", resp.ErrDescription)
}

// ------------------------------------------------------------------------------------------------------------------------
// HandleDbError

func newPgError(code, constraintName, detail, message, columnName string) *pgconn.PgError {
	return &pgconn.PgError{
		Code:           code,
		ConstraintName: constraintName,
		Detail:         detail,
		Message:        message,
		ColumnName:     columnName,
	}
}

func TestHandleDbError_CheckViolation(t *testing.T) {
	w := httptest.NewRecorder()
	pgErr := newPgError(pgerrcode.CheckViolation, "age_positive", "", "", "")
	HandleDbError(context.Background(), 0, "", fmt.Errorf("db error: %w", pgErr), discardLog, w)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := decodeStdResponse(t, w)
	assert.Contains(t, resp.ErrDescription, "check constraint violation")
	assert.Contains(t, resp.ErrDescription, "age_positive")
}

func TestHandleDbError_UniqueViolation_ConstraintName(t *testing.T) {
	w := httptest.NewRecorder()
	pgErr := newPgError(pgerrcode.UniqueViolation, "users_email_key", "", "", "")
	HandleDbError(context.Background(), 0, "", fmt.Errorf("db error: %w", pgErr), discardLog, w)

	assert.Equal(t, http.StatusConflict, w.Code)
	resp := decodeStdResponse(t, w)
	assert.Contains(t, resp.ErrDescription, "unique constraint violation")
	assert.Contains(t, resp.ErrDescription, "users_email_key")
}

func TestHandleDbError_UniqueViolation_Detail(t *testing.T) {
	// Detail is preferred over ConstraintName when present
	w := httptest.NewRecorder()
	pgErr := newPgError(pgerrcode.UniqueViolation, "users_email_name_key", "Key (email, name)=(a@b.com, bob) already exists.", "", "")
	HandleDbError(context.Background(), 0, "", fmt.Errorf("db error: %w", pgErr), discardLog, w)

	assert.Equal(t, http.StatusConflict, w.Code)
	resp := decodeStdResponse(t, w)
	assert.Contains(t, resp.ErrDescription, "unique constraint violation")
	assert.Contains(t, resp.ErrDescription, "Key (email, name)")
}

func TestHandleDbError_ForeignKeyViolation(t *testing.T) {
	w := httptest.NewRecorder()
	pgErr := newPgError(pgerrcode.ForeignKeyViolation, "", "Key (org_id)=(99) is not present in table \"orgs\".", "", "")
	HandleDbError(context.Background(), 0, "", fmt.Errorf("db error: %w", pgErr), discardLog, w)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := decodeStdResponse(t, w)
	assert.Contains(t, resp.ErrDescription, "foreign key violation")
}

func TestHandleDbError_ExclusionViolation(t *testing.T) {
	w := httptest.NewRecorder()
	pgErr := newPgError(pgerrcode.ExclusionViolation, "", "conflicting range detail", "", "")
	HandleDbError(context.Background(), 0, "", fmt.Errorf("db error: %w", pgErr), discardLog, w)

	assert.Equal(t, http.StatusConflict, w.Code)
	resp := decodeStdResponse(t, w)
	assert.Contains(t, resp.ErrDescription, "exclusion constraint violation")
}

func TestHandleDbError_InvalidTextRepresentation(t *testing.T) {
	w := httptest.NewRecorder()
	pgErr := newPgError(pgerrcode.InvalidTextRepresentation, "", "", "invalid input value for enum status: \"unknown\"", "")
	HandleDbError(context.Background(), 0, "", fmt.Errorf("db error: %w", pgErr), discardLog, w)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := decodeStdResponse(t, w)
	assert.Contains(t, resp.ErrDescription, "invalid text")
}

func TestHandleDbError_UnknownPgError(t *testing.T) {
	// an unrecognised pg error code should return a generic 500
	w := httptest.NewRecorder()
	pgErr := newPgError(pgerrcode.DiskFull, "", "", "disk full", "")
	HandleDbError(context.Background(), 0, "", fmt.Errorf("db error: %w", pgErr), discardLog, w)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := decodeStdResponse(t, w)
	assert.Contains(t, resp.ErrDescription, "database error occurred")
}

func TestHandleDbError_LineIncluded(t *testing.T) {
	// when a line number is provided it should appear in the user-facing message
	w := httptest.NewRecorder()
	pgErr := newPgError(pgerrcode.CheckViolation, "age_positive", "", "", "")
	HandleDbError(context.Background(), 3, "INSERT ...", fmt.Errorf("db error: %w", pgErr), discardLog, w)

	resp := decodeStdResponse(t, w)
	assert.Contains(t, resp.ErrDescription, "line 3")
}

func TestHandleDbError_NonPgError(t *testing.T) {
	// a plain error (no pg wrapping) should return a generic 500
	w := httptest.NewRecorder()
	HandleDbError(context.Background(), 0, "SELECT 1", errors.New("connection reset"), discardLog, w)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := decodeStdResponse(t, w)
	assert.Contains(t, resp.ErrDescription, "database error occurred")
}

// ------------------------------------------------------------------------------------------------------------------------
// HandleError

func TestHandleError_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	w := httptest.NewRecorder()
	HandleError(ctx, context.Canceled, discardLog, w)

	// canceled request: no response written
	assert.Equal(t, http.StatusOK, w.Code) // recorder default, nothing written
	assert.Empty(t, w.Body.String())
}

func TestHandleError_ErrNoRows(t *testing.T) {
	w := httptest.NewRecorder()
	HandleError(context.Background(), pgx.ErrNoRows, discardLog, w)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := decodeStdResponse(t, w)
	assert.Equal(t, "row(s) not found", resp.ErrDescription)
}

func TestHandleError_ErrTooManyRows(t *testing.T) {
	w := httptest.NewRecorder()
	HandleError(context.Background(), pgx.ErrTooManyRows, discardLog, w)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := decodeStdResponse(t, w)
	assert.Equal(t, "too many rows found", resp.ErrDescription)
}

func TestHandleError_UserError(t *testing.T) {
	w := httptest.NewRecorder()
	userErr := lyserr.User{Message: "invalid field value"}
	HandleError(context.Background(), userErr, discardLog, w)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := decodeStdResponse(t, w)
	assert.Equal(t, "invalid field value", resp.ErrDescription)
}

func TestHandleError_WrappedUserError(t *testing.T) {
	w := httptest.NewRecorder()
	userErr := lyserr.User{Message: "wrapped user error"}
	wrapped := fmt.Errorf("outer: %w", userErr)
	HandleError(context.Background(), wrapped, discardLog, w)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := decodeStdResponse(t, w)
	assert.Equal(t, "wrapped user error", resp.ErrDescription)
}

func TestHandleError_ExtError(t *testing.T) {
	w := httptest.NewRecorder()
	extErr := lyserr.Ext{Message: "payment gateway error", Err: errors.New("gateway timeout")}
	HandleError(context.Background(), extErr, discardLog, w)

	assert.Equal(t, http.StatusBadGateway, w.Code)
	resp := decodeStdResponse(t, w)
	assert.Equal(t, "payment gateway error", resp.ErrDescription)
}

func TestHandleError_DbError_UserFacing(t *testing.T) {
	pgErr := newPgError(pgerrcode.UniqueViolation, "users_email_key", "", "", "")
	dbErr := lyserr.Db{Err: fmt.Errorf("db: %w", pgErr), Stmt: "INSERT INTO users ..."}
	w := httptest.NewRecorder()
	HandleError(context.Background(), dbErr, discardLog, w)

	assert.Equal(t, http.StatusConflict, w.Code)
	resp := decodeStdResponse(t, w)
	assert.Contains(t, resp.ErrDescription, "unique constraint violation")
}

func TestHandleError_DbError_Internal(t *testing.T) {
	dbErr := lyserr.Db{Err: errors.New("connection lost"), Stmt: "SELECT 1"}
	w := httptest.NewRecorder()
	HandleError(context.Background(), dbErr, discardLog, w)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := decodeStdResponse(t, w)
	assert.Contains(t, resp.ErrDescription, "database error occurred")
}

func TestHandleError_UnknownError(t *testing.T) {
	w := httptest.NewRecorder()
	HandleError(context.Background(), errors.New("something unexpected"), discardLog, w)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := decodeStdResponse(t, w)
	assert.Equal(t, "An internal error occurred", resp.ErrDescription)
}
