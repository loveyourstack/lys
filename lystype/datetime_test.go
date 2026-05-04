package lystype

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDatetimeSuccess(t *testing.T) {

	datetimeStr := "\"2024-06-15 14:30:00+01\""
	var dt Datetime

	// test unmarshal
	err := dt.UnmarshalJSON([]byte(datetimeStr))
	assert.NoError(t, err, "UnmarshalJSON should not error")

	// test marshal
	marshalled, err := dt.MarshalJSON()
	assert.NoError(t, err, "MarshalJSON should not error")
	assert.Equal(t, datetimeStr, string(marshalled), "marshalled datetime")

	// test Scan
	var dt2 Datetime
	err = dt2.Scan(time.Date(2024, 6, 15, 14, 30, 0, 0, time.FixedZone("Europe/Berlin", 1*60*60)))
	assert.NoError(t, err, "Scan should not error")

	marshalled2, err := dt2.MarshalJSON()
	assert.NoError(t, err, "MarshalJSON should not error")
	assert.Equal(t, datetimeStr, string(marshalled2), "marshalled datetime after Scan")
}

func TestDatetimeUnmarshalNull(t *testing.T) {
	var dt Datetime
	err := dt.UnmarshalJSON([]byte("null"))
	assert.NoError(t, err)
	assert.True(t, dt.IsZero())
}

func TestDatetimeUnmarshalInvalid(t *testing.T) {
	var dt Datetime
	err := dt.UnmarshalJSON([]byte("\"2024-06-15T14:30:00Z\""))
	assert.Error(t, err)
}

func TestDatetimeScanNilAndUnsupported(t *testing.T) {
	var dt Datetime
	err := dt.Scan(nil)
	assert.Error(t, err)

	err = dt.Scan("2024-06-15 14:30:00+01")
	assert.Error(t, err)
}

func TestDatetimeStringAndToTime(t *testing.T) {
	var dt Datetime
	err := dt.UnmarshalJSON([]byte("\"2024-06-15 14:30:00+01\""))
	assert.NoError(t, err)

	assert.Equal(t, "2024-06-15 14:30:00+01", dt.String())
	assert.Equal(t, "2024-06-15 14:30:00+01", dt.ToTime().Format(DatetimeFormat))
}
