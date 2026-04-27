package lyspg

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/exp/constraints"
)

// error strings used by internal errors
const (
	ErrDescInsertScanFailed      string = "db.QueryRow or Scan failed"
	ErrDescUpdateExecFailed      string = "db.Exec failed"
	ErrDescGetRowsAffectedFailed string = "sqlRes.RowsAffected() failed"
)

// max number of characters of a statement to print in error logs
const MaxStmtPrintChars int = 5000

// PoolOrTx is an abstraction of a pgx connection pool or transaction, e.g. pgxpool.Pool, pgx.Conn or pgx.Tx
// adapted from Querier in https://github.com/georgysavva/scany/blob/master/pgxscan/pgxscan.go
type PoolOrTx interface {
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) (br pgx.BatchResults)
}

// PrimaryKeyType defines the type constraint of DB primary keys
type PrimaryKeyType interface {
	constraints.Integer | uuid.UUID | ~string
}

// TrackingColNames is the list of reserved tracking column names that are automatically set in Store operations
var TrackingColNames = []string{"created_at", "created_by", "updated_at", "last_user_update_by"}
