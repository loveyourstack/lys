package coretypetest

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	schemaName     string = "core"
	tableName      string = "type_test"
	viewName       string = "type_test"
	pkColName      string = "id"
	defaultOrderBy string = "id"
)

type Input struct {
	CBool      bool                 `db:"c_bool" json:"c_bool"`
	CBoolN     *bool                `db:"c_booln" json:"c_booln"`
	CBoolA     []bool               `db:"c_boola" json:"c_boola"` // array types: if zero, must be set to empty array if db col is not null
	CInt       int64                `db:"c_int" json:"c_int"`
	CIntN      *int64               `db:"c_intn" json:"c_intn"`
	CIntA      []int64              `db:"c_inta" json:"c_inta"`
	CDouble    float32              `db:"c_double" json:"c_double"`
	CDoubleN   *float32             `db:"c_doublen" json:"c_doublen"`
	CDoubleA   []float32            `db:"c_doublea" json:"c_doublea"`
	CNumeric   float32              `db:"c_numeric" json:"c_numeric"`
	CNumericN  *float32             `db:"c_numericn" json:"c_numericn"`
	CNumericA  []float32            `db:"c_numerica" json:"c_numerica"`
	CDate      lystype.Date         `db:"c_date" json:"c_date"`
	CDateN     *lystype.Date        `db:"c_daten" json:"c_daten"`
	CDateA     []pgtype.Date        `db:"c_datea" json:"c_datea"`
	CTime      lystype.Time         `db:"c_time" json:"c_time"`
	CTimeN     *lystype.Time        `db:"c_timen" json:"c_timen"`
	CTimeA     []pgtype.Time        `db:"c_timea" json:"c_timea"`
	CDatetime  lystype.Datetime     `db:"c_datetime" json:"c_datetime"`
	CDatetimeN *lystype.Datetime    `db:"c_datetimen" json:"c_datetimen"`
	CDatetimeA []pgtype.Timestamptz `db:"c_datetimea" json:"c_datetimea"` // same timezone used on entry, but entry 1 stored as local timezone, entry 2 as specified
	CEnum      string               `db:"c_enum" json:"c_enum"`           // enum has no zero value: must be set
	CEnumN     *string              `db:"c_enumn" json:"c_enumn"`
	CEnumA     []string             `db:"c_enuma" json:"c_enuma"`
	CText      string               `db:"c_text" json:"c_text"`
	CTextN     *string              `db:"c_textn" json:"c_textn"`
	CTextA     []string             `db:"c_texta" json:"c_texta"`
}

//CDate  pgtype.Date  `db:"c_date" json:"c_date"` // has no zero value: must be set
//CTime  pgtype.Time   `db:"c_time" json:"c_time"` // json: marshals to struct with Microseconds and Valid
//CDatetime  pgtype.Timestamptz  `db:"c_datetime" json:"c_datetime"` // has no zero value: must be set

// lystype types: work for output, not for input (cannot find encode plan)
// using pgtype for these until fixed
//CDatea    []lystype.Date `db:"c_datea" json:"c_datea"` // works for output, not for input (cannot find encode plan)
//CTimea []lystype.Time `db:"c_timea" json:"c_timea"`
//CDatetimea []lystype.Datetime `db:"c_datetimea" json:"c_datetimea"`

type Model struct {
	Id   int64     `db:"id" json:"id"`
	Iduu uuid.UUID `db:"id_uu" json:"id_uu"`
	Input
}

var (
	meta, inputMeta lysmeta.Result
)

func init() {
	var err error
	meta, err = lysmeta.AnalyzeStructs(reflect.ValueOf(&Input{}).Elem(), reflect.ValueOf(&Model{}).Elem())
	if err != nil {
		log.Fatalf("lysmeta.AnalyzeStructs failed for %s.%s: %s", schemaName, tableName, err.Error())
	}
	inputMeta, _ = lysmeta.AnalyzeStructs(reflect.ValueOf(&Input{}).Elem())
}

type Store struct {
	Db *pgxpool.Pool
}

func (s Store) Delete(ctx context.Context, id int64) (stmt string, err error) {
	return lyspg.DeleteUnique(ctx, s.Db, schemaName, tableName, pkColName, id)
}

func GetEmptyInput() (input Input) {

	// arrays and enum need values, others don't
	return Input{
		CBoolA:     []bool{},
		CIntA:      []int64{},
		CDoubleA:   []float32{},
		CNumericA:  []float32{},
		CDateA:     []pgtype.Date{},
		CTimeA:     []pgtype.Time{},
		CDatetimeA: []pgtype.Timestamptz{},
		CEnum:      "Monday",
		CEnumA:     []string{},
		CTextA:     []string{},
	}
}

func GetFilledInput() (input Input, err error) {

	// date
	d1, err := time.Parse(lystype.DateFormat, "2001-02-03")
	if err != nil {
		return Input{}, fmt.Errorf("time.Parse (d1) failed: %w", err)
	}
	d2, err := time.Parse(lystype.DateFormat, "2002-03-04")
	if err != nil {
		return Input{}, fmt.Errorf("time.Parse (d2) failed: %w", err)
	}

	// time
	t1, err := time.Parse(lystype.TimeFormat, "12:01")
	if err != nil {
		return Input{}, fmt.Errorf("time.Parse (t1) failed: %w", err)
	}
	t2, err := time.Parse(lystype.TimeFormat, "12:02")
	if err != nil {
		return Input{}, fmt.Errorf("time.Parse (t2) failed: %w", err)
	}

	// datetime
	dt1, err := time.Parse(lystype.DatetimeFormat, "2001-02-03 12:01:00+01")
	if err != nil {
		return Input{}, fmt.Errorf("time.Parse (dt1) failed: %w", err)
	}
	dt2, err := time.Parse(lystype.DatetimeFormat, "2002-03-04 12:02:00+01")
	if err != nil {
		return Input{}, fmt.Errorf("time.Parse (dt2) failed: %w", err)
	}

	input = Input{
		CBool:  false,
		CBoolN: lystype.ToPtr(true),
		CBoolA: []bool{false, true},

		CInt:  1,
		CIntN: lystype.ToPtr(int64(2)),
		CIntA: []int64{1, 2},

		CDouble:  1.1,
		CDoubleN: lystype.ToPtr(float32(2.1)),
		CDoubleA: []float32{1.1, 2.1},

		CNumeric:  1.11,
		CNumericN: lystype.ToPtr(float32(2.11)),
		CNumericA: []float32{1.11, 2.11},

		CDate:  lystype.Date(d1),
		CDateN: lystype.ToPtr(lystype.Date(d2)),
		CDateA: []pgtype.Date{}, // TODO

		CTime:  lystype.Time(t1),
		CTimeN: lystype.ToPtr(lystype.Time(t2)),
		CTimeA: []pgtype.Time{}, // TODO

		CDatetime:  lystype.Datetime(dt1),
		CDatetimeN: lystype.ToPtr(lystype.Datetime(dt2)),
		CDatetimeA: []pgtype.Timestamptz{}, // TODO

		CEnum:  "Monday",
		CEnumN: lystype.ToPtr("Tuesday"),
		CEnumA: []string{"Monday", "Tuesday"},

		CText:  "a b",
		CTextN: lystype.ToPtr("b c"),
		CTextA: []string{"a b", "b c"},
	}

	return input, nil
}

func (s Store) GetJsonFields() []string {
	return meta.JsonTags
}

func (s Store) Insert(ctx context.Context, input Input) (newItem Model, stmt string, err error) {
	return lyspg.Insert[Input, Model](ctx, s.Db, schemaName, tableName, viewName, pkColName, meta.DbTags, input)
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, stmt string, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, meta.DbTags, params)
}

func (s Store) SelectById(ctx context.Context, fields []string, id int64) (item Model, stmt string, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, fields, meta.DbTags, id)
}

func (s Store) SelectByUuid(ctx context.Context, fields []string, id uuid.UUID) (item Model, stmt string, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, "id_uu", fields, meta.DbTags, id)
}

func (s Store) Update(ctx context.Context, input Input, id int64) (stmt string, err error) {
	return lyspg.Update(ctx, s.Db, schemaName, tableName, pkColName, input, id)
}

func (s Store) UpdatePartial(ctx context.Context, assignmentsMap map[string]any, id int64) (stmt string, err error) {
	return lyspg.UpdatePartial(ctx, s.Db, schemaName, tableName, pkColName, inputMeta.DbTags, assignmentsMap, id)
}

func (s Store) Validate(validate *validator.Validate, input Input) error {
	return lysmeta.Validate[Input](validate, input)
}
