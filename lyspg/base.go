package lyspg

import (
	"context"
	"reflect"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/loveyourstack/lys/lystype"
	"golang.org/x/exp/constraints"
)

func init() {
	// reflect lystype date types one time on startup for use in getInputValue
	gLystypeDateType = reflect.TypeOf((lystype.Date{}))
	gLystypeDateTypeP = reflect.TypeOf((*lystype.Date)(nil))
	gLystypeTimeType = reflect.TypeOf(lystype.Time{})
	gLystypeTimeTypeP = reflect.TypeOf((*lystype.Time)(nil))
	gLystypeDatetimeType = reflect.TypeOf((lystype.Datetime{}))
	gLystypeDatetimeTypeP = reflect.TypeOf((*lystype.Datetime)(nil))
}

var (
	gLystypeDateType      reflect.Type
	gLystypeDateTypeP     reflect.Type
	gLystypeTimeType      reflect.Type
	gLystypeTimeTypeP     reflect.Type
	gLystypeDatetimeType  reflect.Type
	gLystypeDatetimeTypeP reflect.Type
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
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// PrimaryKeyType defines the type constraint of DB primary keys
type PrimaryKeyType interface {
	constraints.Integer | uuid.UUID | ~string
}

// getInputValsFromStruct returns input values for pg operations from the supplied reflected struct variable
func getInputValsFromStruct(reflVal reflect.Value, omitDbTags []string) (inputVals []any) {

	reflType := reflVal.Type()

	// for each struct field
	for i := 0; i < reflVal.NumField(); i++ {

		// if there are omissions
		if len(omitDbTags) > 0 {

			// get field db tag
			typeField := reflVal.Type().Field(i)
			tag := typeField.Tag
			dbTag := tag.Get("db")

			// skip if dbTag should be omitted
			if dbTag != "" && slices.Contains(omitDbTags, dbTag) {
				continue
			}
		}

		//fmt.Printf("Index: %d\tName: %s\tType: %s\tValue: %v\n", reflType.Field(i).Index[0], reflType.Field(i).Name, reflType.Field(i).Type, reflVal.Field(i))

		// append the input value from the field value and type
		inputVals = append(inputVals, getInputValue(reflVal.Field(i).Interface(), reflType.Field(i).Type))
	}

	return inputVals
}

// getInputValue returns val, but contains special handling for both the pointer and non-pointer versions of lystype date types
func getInputValue(val any, reflType reflect.Type) (inputVal any) {

	switch reflType {

	// lystype.Date
	case gLystypeDateType:
		val := val.(lystype.Date)
		return val.Format(lystype.DateFormat)
	case gLystypeDateTypeP:
		val := val.(*lystype.Date)
		if val == nil {
			return nil
		}
		return (*val).Format(lystype.DateFormat)

	// lystype.Time
	case gLystypeTimeType:
		val := val.(lystype.Time)
		return val.Format(lystype.TimeFormatDb)
	case gLystypeTimeTypeP:
		val := val.(*lystype.Time)
		if val == nil {
			return nil
		}
		return (*val).Format(lystype.TimeFormatDb)

	// lystype.Datetime
	case gLystypeDatetimeType:
		val := val.(lystype.Datetime)
		return val.Format(lystype.DatetimeFormat)
	case gLystypeDatetimeTypeP:
		val := val.(*lystype.Datetime)
		if val == nil {
			return nil
		}
		return (*val).Format(lystype.DatetimeFormat)

	default:
		return val
	}
}
