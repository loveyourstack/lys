package lystype

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDate(t *testing.T) {

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
