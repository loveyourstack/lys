package lyspg

import (
	"reflect"
	"testing"
	"time"

	"github.com/loveyourstack/lys/lystype"
	"github.com/stretchr/testify/assert"
)

func TestGetInputValsFromStruct_FlatStruct(t *testing.T) {

	type input struct {
		Name string `db:"name"`
		Age  int    `db:"age"`
	}

	in := input{Name: "Alice", Age: 30}
	vals := getInputValsFromStruct(reflect.ValueOf(in), nil)

	assert.Equal(t, 2, len(vals))
	assert.Equal(t, "Alice", vals[0])
	assert.Equal(t, 30, vals[1])
}

func TestGetInputValsFromStruct_NestedStruct(t *testing.T) {

	type inner struct {
		Name string `db:"name"`
		Age  int    `db:"age"`
	}
	type outer struct {
		Inner     inner
		CreatedBy string `db:"created_by"`
	}

	in := outer{
		Inner:     inner{Name: "Bob", Age: 25},
		CreatedBy: "admin",
	}
	vals := getInputValsFromStruct(reflect.ValueOf(in), nil)

	assert.Equal(t, 3, len(vals))
	assert.Equal(t, "Bob", vals[0])
	assert.Equal(t, 25, vals[1])
	assert.Equal(t, "admin", vals[2])
}

func TestGetInputValsFromStruct_OmitDbTags(t *testing.T) {

	type input struct {
		Name string `db:"name"`
		Age  int    `db:"age"`
		City string `db:"city"`
	}

	in := input{Name: "Carol", Age: 40, City: "Berlin"}

	// omit age
	vals := getInputValsFromStruct(reflect.ValueOf(in), []string{"age"})

	assert.Equal(t, 2, len(vals))
	assert.Equal(t, "Carol", vals[0])
	assert.Equal(t, "Berlin", vals[1])
}

func TestGetInputValsFromStruct_LystypeDate(t *testing.T) {

	type input struct {
		Name    string       `db:"name"`
		StartDt lystype.Date `db:"start_dt"`
	}

	dt, err := time.Parse(lystype.DateFormat, "2024-03-15")
	assert.NoError(t, err)

	in := input{Name: "Dave", StartDt: lystype.Date(dt)}
	vals := getInputValsFromStruct(reflect.ValueOf(in), nil)

	assert.Equal(t, 2, len(vals))
	assert.Equal(t, "Dave", vals[0])
	assert.Equal(t, "2024-03-15", vals[1])
}

func TestGetInputValsFromStruct_LystypeTime(t *testing.T) {

	type input struct {
		Label   string       `db:"label"`
		StartTm lystype.Time `db:"start_tm"`
	}

	tm, err := time.Parse(lystype.TimeFormat, "14:30")
	assert.NoError(t, err)

	in := input{Label: "morning", StartTm: lystype.Time(tm)}
	vals := getInputValsFromStruct(reflect.ValueOf(in), nil)

	assert.Equal(t, 2, len(vals))
	assert.Equal(t, "morning", vals[0])
	assert.Equal(t, "14:30:00", vals[1])
}

func TestGetInputValsFromStruct_LystypeDatetime(t *testing.T) {

	type input struct {
		Label   string           `db:"label"`
		EventDt lystype.Datetime `db:"event_dt"`
	}

	dt, err := time.Parse(lystype.DatetimeFormat, "2024-03-15 09:30:00+01")
	assert.NoError(t, err)

	in := input{Label: "event", EventDt: lystype.Datetime(dt)}
	vals := getInputValsFromStruct(reflect.ValueOf(in), nil)

	assert.Equal(t, 2, len(vals))
	assert.Equal(t, "event", vals[0])
	assert.Equal(t, "2024-03-15 09:30:00+01", vals[1])
}

func TestGetInputValsFromStruct_LystypeDatePointer(t *testing.T) {

	type input struct {
		Name    string        `db:"name"`
		StartDt *lystype.Date `db:"start_dt"`
	}

	dt, err := time.Parse(lystype.DateFormat, "2024-06-01")
	assert.NoError(t, err)
	d := lystype.Date(dt)

	// non-nil pointer
	in := input{Name: "Eve", StartDt: &d}
	vals := getInputValsFromStruct(reflect.ValueOf(in), nil)
	assert.Equal(t, 2, len(vals))
	assert.Equal(t, "Eve", vals[0])
	assert.Equal(t, "2024-06-01", vals[1])

	// nil pointer
	in2 := input{Name: "Eve", StartDt: nil}
	vals2 := getInputValsFromStruct(reflect.ValueOf(in2), nil)
	assert.Equal(t, 2, len(vals2))
	assert.Equal(t, "Eve", vals2[0])
	assert.Nil(t, vals2[1])
}

func TestGetInputValsFromStruct_NestedWithLystype(t *testing.T) {

	type inner struct {
		Name    string       `db:"name"`
		StartDt lystype.Date `db:"start_dt"`
	}
	type outer struct {
		Inner     inner
		CreatedBy string `db:"created_by"`
	}

	dt, err := time.Parse(lystype.DateFormat, "2024-01-01")
	assert.NoError(t, err)

	in := outer{
		Inner:     inner{Name: "Frank", StartDt: lystype.Date(dt)},
		CreatedBy: "system",
	}
	vals := getInputValsFromStruct(reflect.ValueOf(in), nil)

	assert.Equal(t, 3, len(vals))
	assert.Equal(t, "Frank", vals[0])
	assert.Equal(t, "2024-01-01", vals[1])
	assert.Equal(t, "system", vals[2])
}

func TestGetInputValsFromStruct_NestedOmitDbTags(t *testing.T) {

	type inner struct {
		Name string `db:"name"`
		Age  int    `db:"age"`
	}
	type outer struct {
		Inner     inner
		CreatedBy string `db:"created_by"`
	}

	in := outer{
		Inner:     inner{Name: "Grace", Age: 50},
		CreatedBy: "admin",
	}
	vals := getInputValsFromStruct(reflect.ValueOf(in), []string{"age"})

	assert.Equal(t, 2, len(vals))
	assert.Equal(t, "Grace", vals[0])
	assert.Equal(t, "admin", vals[1])
}

func TestGetInputValsFromStruct_EmptyStruct(t *testing.T) {

	type input struct{}

	in := input{}
	vals := getInputValsFromStruct(reflect.ValueOf(in), nil)

	assert.Equal(t, 0, len(vals))
}
