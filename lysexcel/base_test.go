package lysexcel

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
	"time"

	"codeberg.org/tealeg/xlsx/v4"
	"github.com/loveyourstack/lys/lystype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type writeItemsRec struct {
	CBool     bool             `json:"c_bool"`
	CDate     lystype.Date     `json:"c_date"`
	CDatetime lystype.Datetime `json:"c_datetime"`
	CFloat64  float64          `json:"c_float64"`
	CInt64    int64            `json:"c_int64"`
	CString   string           `json:"c_string"`
	CTime     lystype.Time     `json:"c_time"`
	Ignored   string           `json:"-"`
	NoJSONTag string
}

func TestWriteItemsSuccess(t *testing.T) {
	timeDate := time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC)
	timeDT := time.Date(2026, 4, 23, 12, 34, 56, 0, time.FixedZone("UTC+2", 2*3600))
	timeOnly := time.Date(2026, 4, 23, 9, 15, 0, 0, time.UTC)

	items := []writeItemsRec{{
		CBool:     true,
		CDate:     lystype.Date(timeDate),
		CDatetime: lystype.Datetime(timeDT),
		CFloat64:  1.5,
		CInt64:    42,
		CString:   "alpha",
		CTime:     lystype.Time(timeOnly),
		Ignored:   "x",
		NoJSONTag: "y",
	}}

	jsonTagTypeMap := map[string]reflect.Type{
		"c_bool":     reflect.TypeFor[bool](),
		"c_date":     reflect.TypeFor[lystype.Date](),
		"c_datetime": reflect.TypeFor[lystype.Datetime](),
		"c_float64":  reflect.TypeFor[float64](),
		"c_int64":    reflect.TypeFor[int64](),
		"c_string":   reflect.TypeFor[string](),
		"c_time":     reflect.TypeFor[lystype.Time](),
	}

	var b bytes.Buffer
	err := WriteItems(items, jsonTagTypeMap, "", &b)
	require.NoError(t, err)
	require.NotZero(t, b.Len())

	wb, err := xlsx.OpenBinary(b.Bytes())
	require.NoError(t, err)
	require.NotEmpty(t, wb.Sheets)

	sh := wb.Sheets[0]
	assert.Equal(t, "data", sh.Name)
	require.GreaterOrEqual(t, sh.MaxRow, 2)

	headerRow, err := sh.Row(0)
	require.NoError(t, err)
	dataRow, err := sh.Row(1)
	require.NoError(t, err)

	assert.Equal(t, "c_bool", headerRow.GetCell(0).String())
	assert.Equal(t, "c_date", headerRow.GetCell(1).String())
	assert.Equal(t, "c_datetime", headerRow.GetCell(2).String())
	assert.Equal(t, "c_float64", headerRow.GetCell(3).String())
	assert.Equal(t, "c_int64", headerRow.GetCell(4).String())
	assert.Equal(t, "c_string", headerRow.GetCell(5).String())
	assert.Equal(t, "c_time", headerRow.GetCell(6).String())

	assert.Equal(t, "TRUE", dataRow.GetCell(0).String())
	assert.Equal(t, "04-23-26", dataRow.GetCell(1).String())      // defaults to US format
	assert.Equal(t, "4/23/26 10:34", dataRow.GetCell(2).String()) // defaults to US format and UTC
	assert.Equal(t, "1.5", dataRow.GetCell(3).String())
	assert.Equal(t, "42", dataRow.GetCell(4).String())
	assert.Equal(t, "alpha", dataRow.GetCell(5).String())
	assert.Equal(t, "09:15", dataRow.GetCell(6).String())
}

func TestWriteItemsFailureValidation(t *testing.T) {
	t.Run("empty items", func(t *testing.T) {
		var b bytes.Buffer
		err := WriteItems([]writeItemsRec{}, map[string]reflect.Type{"c_string": reflect.TypeFor[string]()}, "", &b)
		assert.EqualError(t, err, "items is empty")
	})

	t.Run("empty jsonTagTypeMap", func(t *testing.T) {
		var b bytes.Buffer
		err := WriteItems([]writeItemsRec{{CString: "a"}}, map[string]reflect.Type{}, "", &b)
		assert.EqualError(t, err, "jsonTagTypeMap is empty")
	})

	t.Run("nil writer", func(t *testing.T) {
		err := WriteItems([]writeItemsRec{{CString: "a"}}, map[string]reflect.Type{"c_string": reflect.TypeFor[string]()}, "", nil)
		assert.EqualError(t, err, "writer is mandatory")
	})
}

type errWriter struct{}

func (w errWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write failed")
}

func TestWriteItemsFailureWrite(t *testing.T) {
	items := []writeItemsRec{{CString: "a"}}
	jsonTagTypeMap := map[string]reflect.Type{"c_string": reflect.TypeFor[string]()}

	err := WriteItems(items, jsonTagTypeMap, "", errWriter{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "wb.Write failed")
	assert.Contains(t, err.Error(), "write failed")
}
