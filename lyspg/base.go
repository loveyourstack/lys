package lyspg

import (
	"context"
	"reflect"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lystype"
	"golang.org/x/exp/constraints"
)

func init() {
	// reflect lystype date types one time on startup for use in getInputValue
	gLystypeDateType = reflect.TypeOf((lystype.Date{}))
	gLystypeDateTypeP = reflect.TypeOf((*lystype.Date)(nil))
	gLystypeDateTypeA = reflect.TypeOf([]lystype.Date{})

	gLystypeTimeType = reflect.TypeOf(lystype.Time{})
	gLystypeTimeTypeP = reflect.TypeOf((*lystype.Time)(nil))
	gLystypeTimeTypeA = reflect.TypeOf([]lystype.Time{})

	gLystypeDatetimeType = reflect.TypeOf((lystype.Datetime{}))
	gLystypeDatetimeTypeP = reflect.TypeOf((*lystype.Datetime)(nil))
	gLystypeDatetimeTypeA = reflect.TypeOf([]lystype.Datetime{})
}

var (
	gLystypeDateType  reflect.Type
	gLystypeDateTypeP reflect.Type
	gLystypeDateTypeA reflect.Type

	gLystypeTimeType  reflect.Type
	gLystypeTimeTypeP reflect.Type
	gLystypeTimeTypeA reflect.Type

	gLystypeDatetimeType  reflect.Type
	gLystypeDatetimeTypeP reflect.Type
	gLystypeDatetimeTypeA reflect.Type
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

// getInputValsFromStruct returns input values for pg operations from the supplied reflected struct variable
func getInputValsFromStruct(reflVal reflect.Value, omitDbTags []string) (inputVals []any) {

	reflType := reflVal.Type()

	// for each struct field
	for i := 0; i < reflVal.NumField(); i++ {

		field := reflType.Field(i)

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

		// if this field is a struct (embedded or named) and has db or json tags (omits structs like time.Time that would cause a panic due to unexported fields)
		if field.Type.Kind() == reflect.Struct && lysmeta.HasDbOrJsonTags(field.Type) {

			// recurse into it
			innerInputVals := getInputValsFromStruct(reflVal.Field(i), omitDbTags)
			inputVals = append(inputVals, innerInputVals...)
			continue
		}

		//fmt.Printf("Index: %d\tName: %s\tType: %s\tValue: %v\n", field.Index[0], field.Name, field.Type, reflVal.Field(i))

		// append the input value from the field value and type
		inputVals = append(inputVals, getInputValue(reflVal.Field(i).Interface(), field.Type))
	}

	return inputVals
}

// getInputValue returns val, but contains special handling for lystype date types
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

	case gLystypeDateTypeA:
		val := val.([]lystype.Date)
		inputVal := make([]string, len(val))
		for i, v := range val {
			inputVal[i] = v.Format(lystype.DateFormat)
		}
		return inputVal

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

	case gLystypeTimeTypeA:
		val := val.([]lystype.Time)
		inputVal := make([]string, len(val))
		for i, v := range val {
			inputVal[i] = v.Format(lystype.TimeFormatDb)
		}
		return inputVal

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

	case gLystypeDatetimeTypeA:
		val := val.([]lystype.Datetime)
		inputVal := make([]string, len(val))
		for i, v := range val {
			inputVal[i] = v.Format(lystype.DatetimeFormat)
		}
		return inputVal

	default:
		return val
	}
}
