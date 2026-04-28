package lysmeta

import (
	"reflect"
	"testing"
	"time"

	"github.com/loveyourstack/lys/lystype"
	"github.com/stretchr/testify/assert"
)

func TestGetInputValueRegularTypes(t *testing.T) {

	assert.Equal(t, 1, GetInputValue(1, reflect.TypeFor[int]()))
	assert.Equal(t, "raw", GetInputValue("raw", reflect.TypeFor[string]()))
}

func TestGetInputValueDateTypes(t *testing.T) {

	date := lystype.Date(time.Date(2026, 4, 27, 10, 30, 0, 0, time.UTC))
	assert.Equal(t, "2026-04-27", GetInputValue(date, reflect.TypeFor[lystype.Date]()), "regular value")

	var dateNilPtr *lystype.Date
	assert.Equal(t, nil, GetInputValue(dateNilPtr, reflect.TypeFor[*lystype.Date]()), "nil pointer")

	datePtr := &date
	assert.Equal(t, "2026-04-27", GetInputValue(datePtr, reflect.TypeFor[*lystype.Date]()), "pointer value")

	dtA := []lystype.Date{
		lystype.Date(time.Date(2026, 4, 27, 10, 30, 0, 0, time.UTC)),
		lystype.Date(time.Date(2026, 4, 28, 11, 45, 0, 0, time.UTC)),
	}
	assert.EqualValues(
		t,
		[]string{"2026-04-27", "2026-04-28"},
		GetInputValue(dtA, reflect.TypeFor[[]lystype.Date]()),
		"slice of values",
	)

	dtAEmpty := []lystype.Date{}
	assert.EqualValues(t, []string{}, GetInputValue(dtAEmpty, reflect.TypeFor[[]lystype.Date]()), "empty slice")
}

func TestGetInputValueTimeTypes(t *testing.T) {

	tm := lystype.Time(time.Date(2026, 4, 27, 10, 30, 40, 0, time.UTC))
	assert.Equal(t, "10:30:40", GetInputValue(tm, reflect.TypeFor[lystype.Time]()), "regular value")

	var tmNilPtr *lystype.Time
	assert.Equal(t, nil, GetInputValue(tmNilPtr, reflect.TypeFor[*lystype.Time]()), "nil pointer")

	tmPtr := &tm
	assert.Equal(t, "10:30:40", GetInputValue(tmPtr, reflect.TypeFor[*lystype.Time]()), "pointer value")

	tmA := []lystype.Time{
		lystype.Time(time.Date(2026, 4, 27, 10, 30, 40, 0, time.UTC)),
		lystype.Time(time.Date(2026, 4, 28, 11, 45, 50, 0, time.UTC)),
	}
	assert.EqualValues(
		t,
		[]string{"10:30:40", "11:45:50"},
		GetInputValue(tmA, reflect.TypeFor[[]lystype.Time]()),
		"slice of values",
	)

	tmAEmpty := []lystype.Time{}
	assert.EqualValues(t, []string{}, GetInputValue(tmAEmpty, reflect.TypeFor[[]lystype.Time]()), "empty slice")
}

func TestGetInputValueDatetimeTypes(t *testing.T) {

	dt := lystype.Datetime(time.Date(2026, 4, 27, 10, 30, 40, 0, time.FixedZone("z2", 2*3600)))
	assert.Equal(t, "2026-04-27 10:30:40+02", GetInputValue(dt, reflect.TypeFor[lystype.Datetime]()), "regular value")

	var dtNilPtr *lystype.Datetime
	assert.Equal(t, nil, GetInputValue(dtNilPtr, reflect.TypeFor[*lystype.Datetime]()), "nil pointer")

	dtPtr := &dt
	assert.Equal(t, "2026-04-27 10:30:40+02", GetInputValue(dtPtr, reflect.TypeFor[*lystype.Datetime]()), "pointer value")

	dtA := []lystype.Datetime{
		lystype.Datetime(time.Date(2026, 4, 27, 10, 30, 40, 0, time.UTC)),
		lystype.Datetime(time.Date(2026, 4, 28, 11, 45, 50, 0, time.UTC)),
	}
	assert.EqualValues(
		t,
		[]string{"2026-04-27 10:30:40+00", "2026-04-28 11:45:50+00"},
		GetInputValue(dtA, reflect.TypeFor[[]lystype.Datetime]()),
		"slice of values",
	)

	dtAEmpty := []lystype.Datetime{}
	assert.EqualValues(t, []string{}, GetInputValue(dtAEmpty, reflect.TypeFor[[]lystype.Datetime]()), "empty slice")
}
