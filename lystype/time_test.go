package lystype

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTime(t *testing.T) {

	timeStr := "\"14:30\""
	var ti Time

	// test unmarshal
	err := ti.UnmarshalJSON([]byte(timeStr))
	assert.NoError(t, err, "UnmarshalJSON should not error")

	// test marshal
	marshalled, err := ti.MarshalJSON()
	assert.NoError(t, err, "MarshalJSON should not error")
	assert.Equal(t, timeStr, string(marshalled), "marshalled time")

	// test Scan
	var ti2 Time
	err = ti2.Scan("14:30:00") // db format
	assert.NoError(t, err, "Scan should not error")

	marshalled2, err := ti2.MarshalJSON()
	assert.NoError(t, err, "MarshalJSON should not error")
	assert.Equal(t, timeStr, string(marshalled2), "marshalled time after Scan")
}
