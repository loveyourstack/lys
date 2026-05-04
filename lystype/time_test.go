package lystype

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimeSuccess(t *testing.T) {

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

func TestTimeUnmarshalNull(t *testing.T) {
	var ti Time
	err := ti.UnmarshalJSON([]byte("null"))
	assert.NoError(t, err)
	assert.True(t, ti.IsZero())
}

func TestTimeUnmarshalInvalid(t *testing.T) {
	var ti Time
	err := ti.UnmarshalJSON([]byte("\"14:30:00\""))
	assert.Error(t, err)
}

func TestTimeScanNilAndUnsupported(t *testing.T) {
	var ti Time
	err := ti.Scan(nil)
	assert.Error(t, err)

	err = ti.Scan(143000)
	assert.Error(t, err)
}

func TestTimeScanInvalidFormat(t *testing.T) {
	var ti Time
	err := ti.Scan("14:30")
	assert.Error(t, err)
}

func TestTimeString(t *testing.T) {
	var ti Time
	err := ti.UnmarshalJSON([]byte("\"14:30\""))
	assert.NoError(t, err)

	assert.Equal(t, "14:30", ti.String())
}
