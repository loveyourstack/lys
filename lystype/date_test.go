package lystype

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDateSuccess(t *testing.T) {

	dateStr := "\"2024-06-15\""
	var d Date

	// test unmarshal
	err := d.UnmarshalJSON([]byte(dateStr))
	assert.NoError(t, err, "UnmarshalJSON should not error")

	// test marshal
	marshalled, err := d.MarshalJSON()
	assert.NoError(t, err, "MarshalJSON should not error")
	assert.Equal(t, dateStr, string(marshalled), "marshalled date")

	// test Scan
	var d2 Date
	err = d2.Scan(time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC))
	assert.NoError(t, err, "Scan should not error")

	marshalled2, err := d2.MarshalJSON()
	assert.NoError(t, err, "MarshalJSON should not error")
	assert.Equal(t, dateStr, string(marshalled2), "marshalled date after Scan")
}

func TestDateUnmarshalNull(t *testing.T) {
	var d Date
	err := d.UnmarshalJSON([]byte("null"))
	assert.NoError(t, err)
	assert.True(t, d.IsZero())
}

func TestDateUnmarshalInvalid(t *testing.T) {
	var d Date
	err := d.UnmarshalJSON([]byte("\"2024/06/15\""))
	assert.Error(t, err)
}

func TestDateScanNilAndUnsupported(t *testing.T) {
	var d Date
	err := d.Scan(nil)
	assert.Error(t, err)

	err = d.Scan("2024-06-15")
	assert.Error(t, err)
}

func TestDateStringAndToTime(t *testing.T) {
	var d Date
	err := d.UnmarshalJSON([]byte("\"2024-06-15\""))
	assert.NoError(t, err)

	assert.Equal(t, "2024-06-15", d.String())
	assert.Equal(t, "2024-06-15", d.ToTime().Format(DateFormat))
}
