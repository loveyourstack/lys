package lysgen

import (
	"testing"

	"github.com/loveyourstack/lys/lyspg"
	"github.com/stretchr/testify/assert"
)

func TestGetEqualTypesSuccess(t *testing.T) {

	cols := []lyspg.Column{

		// should be excluded
		{Name: "identity", DataType: "bigint", IsIdentity: true},
		{Name: "generated", DataType: "text", IsGenerated: true},
		{Name: "tracking", DataType: "text", IsTracking: true},

		// should be included
		{Name: "array", DataType: "ARRAY"},
		{Name: "bigint", DataType: "bigint"},
		{Name: "date", DataType: "date"},
		{Name: "numeric", DataType: "numeric"},
		{Name: "text", DataType: "text"},
		{Name: "time", DataType: "time"},
		{Name: "timestamp_with_time_zone", DataType: "timestamp with time zone"},
		{Name: "user_defined", DataType: "USER-DEFINED"}, // enum
	}

	actualA, err := getEqual(cols)
	if err != nil {
		t.Fatalf("getEqual failed: %s", err.Error())
	}

	expectedA := []string{
		"func (s Store) Equal(a, b Model) bool {",
		"    return lysslices.EqualUnordered(a.Array, b.Array) &&",
		"        a.Bigint == b.Bigint &&",
		"        a.Date.Format(lystype.DateFormat) == b.Date.Format(lystype.DateFormat) &&",
		"        fmt.Sprintf(\"%.4f\", a.Numeric) == fmt.Sprintf(\"%.4f\", b.Numeric) &&",
		"        a.Text == b.Text &&",
		"        a.Time.Format(lystype.TimeFormat) == b.Time.Format(lystype.TimeFormat) &&",
		"        a.TimestampWithTimeZone.Format(lystype.DatetimeFormat) == b.TimestampWithTimeZone.Format(lystype.DatetimeFormat) &&",
		"        a.UserDefined == b.UserDefined",
		"}",
	}

	assert.EqualValues(t, expectedA, actualA)
}

func TestGetEqualTypesFailure(t *testing.T) {

	cols := []lyspg.Column{
		{Name: "unknown", DataType: "unknown"},
	}

	_, err := getEqual(cols)
	assert.EqualError(t, err, "no rule type found for DataType: unknown")
}
