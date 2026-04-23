package lyscsv

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
	"time"

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
	err := WriteItems(items, jsonTagTypeMap, ',', &b)
	require.NoError(t, err)

	expected := "c_bool,c_date,c_datetime,c_float64,c_int64,c_string,c_time\n" +
		"true,2026-04-23,2026-04-23 12:34:56+02,1.5,42,alpha,09:15\n"
	assert.Equal(t, expected, b.String())
}

func TestWriteItemsFailureValidation(t *testing.T) {
	t.Run("empty items", func(t *testing.T) {
		var b bytes.Buffer
		err := WriteItems([]writeItemsRec{}, map[string]reflect.Type{"name": reflect.TypeFor[string]()}, ',', &b)
		assert.EqualError(t, err, "items is empty")
	})

	t.Run("empty jsonTagTypeMap", func(t *testing.T) {
		var b bytes.Buffer
		err := WriteItems([]writeItemsRec{{CString: "a"}}, map[string]reflect.Type{}, ',', &b)
		assert.EqualError(t, err, "jsonTagTypeMap is empty")
	})

	t.Run("missing delimiter", func(t *testing.T) {
		var b bytes.Buffer
		err := WriteItems([]writeItemsRec{{CString: "a"}}, map[string]reflect.Type{"c_string": reflect.TypeFor[string]()}, 0, &b)
		assert.EqualError(t, err, "delimiter is mandatory")
	})

	t.Run("nil writer", func(t *testing.T) {
		err := WriteItems([]writeItemsRec{{CString: "a"}}, map[string]reflect.Type{"c_string": reflect.TypeFor[string]()}, ',', nil)
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

	err := WriteItems(items, jsonTagTypeMap, ',', errWriter{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "csv.NewWriter: flush")
	assert.Contains(t, err.Error(), "write failed")
}
