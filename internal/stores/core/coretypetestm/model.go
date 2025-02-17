package coretypetestm

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/loveyourstack/lys/lystype"
	"github.com/stretchr/testify/assert"
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

type Model struct {
	Id   int64     `db:"id" json:"id"`
	Iduu uuid.UUID `db:"id_uu" json:"id_uu"`
	Input
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

// copied from lys: TODO
func mustParseTime(t testing.TB, layout, value string) time.Time {

	timeV, err := time.Parse(layout, value)
	if err != nil {
		t.Fatalf("time.Parse failed: %v", err)
	}

	return timeV
}

func TestEmptyInput(t testing.TB, item Input) {

	// boolean
	assert.EqualValues(t, false, item.CBool, "CBool")
	assert.EqualValues(t, (*bool)(nil), item.CBoolN, "CBoolN")
	assert.EqualValues(t, []bool{}, item.CBoolA, "CBoolA")

	// int
	assert.EqualValues(t, 0, item.CInt, "CInt")
	assert.EqualValues(t, (*int64)(nil), item.CIntN, "CIntN")
	assert.EqualValues(t, []int64{}, item.CIntA, "CIntA")

	// double
	assert.EqualValues(t, float32(0.0), item.CDouble, "CDouble")
	assert.EqualValues(t, (*float32)(nil), item.CDoubleN, "CDoubleN")
	assert.EqualValues(t, []float32{}, item.CDoubleA, "CDoubleA")

	// numeric
	assert.EqualValues(t, float32(0.0), item.CNumeric, "CNumeric")
	assert.EqualValues(t, (*float32)(nil), item.CNumericN, "CNumericN")
	assert.EqualValues(t, []float32{}, item.CNumericA, "CNumericA")

	// date
	assert.EqualValues(t, lystype.Date{}, item.CDate, "CDate")
	assert.EqualValues(t, (*lystype.Date)(nil), item.CDateN, "CDateN")
	// TODO item.CDateA

	// time
	t1 := mustParseTime(t, lystype.TimeFormat, "00:00")
	assert.EqualValues(t, t1, item.CTime, "CTime")
	assert.EqualValues(t, (*lystype.Time)(nil), item.CTimeN, "CTimeN")
	// TODO item.CTimeA

	// datetime
	dt1 := mustParseTime(t, lystype.DatetimeFormat, "0001-01-01 12:00:00+00")
	assert.EqualValues(t, dt1.Format(lystype.DateFormat), item.CDatetime.Format(lystype.DateFormat), "CDatetime")
	assert.EqualValues(t, (*lystype.Datetime)(nil), item.CDatetimeN, "CDatetimeN")
	// TODO item.CDatetimeA

	// enum
	assert.EqualValues(t, "Monday", item.CEnum, "CEnum")
	assert.EqualValues(t, (*string)(nil), item.CEnumN, "CEnumN")
	assert.EqualValues(t, []string{}, item.CEnumA, "CEnumA")

	// text
	assert.EqualValues(t, "", item.CText, "CText")
	assert.EqualValues(t, (*string)(nil), item.CTextN, "CTextN")
	assert.EqualValues(t, []string{}, item.CTextA, "CTextA")
}

func TestFilledInput(t testing.TB, item Input) {

	// boolean
	assert.EqualValues(t, false, item.CBool, "CBool")
	assert.EqualValues(t, true, *item.CBoolN, "CBoolN")
	expectedCBoolA := []bool{false, true}
	for i := range item.CBoolA {
		assert.EqualValues(t, expectedCBoolA[i], item.CBoolA[i], "CBoolA", i)
	}

	// int
	assert.EqualValues(t, int64(1), item.CInt, "CInt")
	assert.EqualValues(t, int64(2), *item.CIntN, "CIntN")
	expectedCIntA := []int64{1, 2}
	for i := range item.CIntA {
		assert.EqualValues(t, expectedCIntA[i], item.CIntA[i], "CIntA", i)
	}

	// double
	assert.EqualValues(t, float32(1.1), item.CDouble, "CDouble")
	assert.EqualValues(t, float32(2.1), *item.CDoubleN, "CDoubleN")
	expectedCDoubleA := []float32{1.1, 2.1}
	for i := range item.CDoubleA {
		assert.EqualValues(t, expectedCDoubleA[i], item.CDoubleA[i], "CDoubleA", i)
	}

	// numeric
	assert.EqualValues(t, float32(1.11), item.CNumeric, "CNumeric")
	assert.EqualValues(t, float32(2.11), *item.CNumericN, "CNumericN")
	expectedCNumericA := []float32{1.11, 2.11}
	for i := range item.CNumericA {
		assert.EqualValues(t, expectedCNumericA[i], item.CNumericA[i], "CNumericA", i)
	}

	// date
	d1 := mustParseTime(t, lystype.DateFormat, "2001-02-03")
	d2 := mustParseTime(t, lystype.DateFormat, "2002-03-04")
	assert.EqualValues(t, lystype.Date(d1), item.CDate, "CDate")
	assert.EqualValues(t, lystype.Date(d2), *item.CDateN, "CDateN")
	// TODO item.CDateA

	// time
	t1 := mustParseTime(t, lystype.TimeFormat, "12:01")
	t2 := mustParseTime(t, lystype.TimeFormat, "12:02")
	assert.EqualValues(t, lystype.Time(t1), item.CTime, "CTime")
	assert.EqualValues(t, lystype.Time(t2), *item.CTimeN, "CTimeN")
	// TODO item.CTimeA

	// datetime
	dt1 := mustParseTime(t, lystype.DatetimeFormat, "2001-02-03 12:01:00+01")
	dt2 := mustParseTime(t, lystype.DatetimeFormat, "2002-03-04 12:02:00+01")
	assert.EqualValues(t, lystype.Datetime(dt1), item.CDatetime, "CDatetime")
	assert.EqualValues(t, lystype.Datetime(dt2), *item.CDatetimeN, "CDatetimeN")
	// TODO item.CDatetimeA

	// enum
	assert.EqualValues(t, "Monday", item.CEnum, "CEnum")
	assert.EqualValues(t, "Tuesday", *item.CEnumN, "CEnumN")
	expectedCEnumA := []string{"Monday", "Tuesday"}
	for i := range item.CEnumA {
		assert.EqualValues(t, expectedCEnumA[i], item.CEnumA[i], "CEnumA", i)
	}

	// text
	assert.EqualValues(t, "a b", item.CText, "CText")
	assert.EqualValues(t, "b c", *item.CTextN, "CTextN")
	expectedCTextA := []string{"a b", "b c"}
	for i := range item.CTextA {
		assert.EqualValues(t, expectedCTextA[i], item.CTextA[i], "CTextA", i)
	}
}
