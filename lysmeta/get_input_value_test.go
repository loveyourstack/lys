package lysmeta

import (
	"reflect"
	"testing"
	"time"

	"github.com/loveyourstack/lys/lystype"
	"github.com/stretchr/testify/assert"
)

func TestGetInputValueSpecialTypes(t *testing.T) {
	date := lystype.Date(time.Date(2026, 4, 27, 10, 30, 0, 0, time.UTC))
	assert.Equal(t, "2026-04-27", GetInputValue(date, reflect.TypeFor[lystype.Date]()))

	var timePtr *lystype.Time
	assert.Nil(t, GetInputValue(timePtr, reflect.TypeFor[*lystype.Time]()))

	dtA := []lystype.Datetime{
		lystype.Datetime(time.Date(2026, 4, 27, 10, 30, 0, 0, time.FixedZone("z1", -5*3600))),
		lystype.Datetime(time.Date(2026, 4, 28, 11, 45, 0, 0, time.FixedZone("z2", 2*3600))),
	}
	assert.EqualValues(
		t,
		[]string{"2026-04-27 10:30:00-05", "2026-04-28 11:45:00+02"},
		GetInputValue(dtA, reflect.TypeFor[[]lystype.Datetime]()),
	)

	assert.Equal(t, "raw", GetInputValue("raw", reflect.TypeFor[string]()))
}
