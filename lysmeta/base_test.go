package lysmeta

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetJsonKey(t *testing.T) {
	assert.Equal(t, "FieldName", getJsonKey("", "FieldName"), "empty tag falls back to field name")
	assert.Equal(t, "FieldName", getJsonKey(",omitempty", "FieldName"), "empty key with options falls back to field name")
	assert.Equal(t, "", getJsonKey("-", "FieldName"), "dash omits key")
	assert.Equal(t, "custom", getJsonKey("custom,omitempty", "FieldName"), "explicit key is used")
}

func TestJsonKeyTypeMap(t *testing.T) {
	plan := Plan{
		Fields: []Field{
			{Name: "A", Type: reflect.TypeFor[string](), JsonKey: "name"},
			{Name: "B", Type: reflect.TypeFor[int64](), JsonKey: "age"},
			{Name: "C", Type: reflect.TypeFor[bool](), JsonKey: ""},
		},
	}

	m := plan.JsonKeyTypeMap()
	assert.EqualValues(t, 2, len(m))
	assert.Equal(t, reflect.TypeFor[string](), m["name"])
	assert.Equal(t, reflect.TypeFor[int64](), m["age"])
	_, exists := m[""]
	assert.False(t, exists)
}

func TestDbValuesSuccess(t *testing.T) {
	type input struct {
		Name string `db:"name_db" json:"name"`
		Age  int64  `db:"age_db" json:"age"`
		Skip string `json:"skip"`
	}

	plan, err := AnalyzeT(input{Name: "james", Age: 42, Skip: "ignored"}, true)
	assert.NoError(t, err)

	dbNames, values, err := plan.DbValues()
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"name_db", "age_db"}, dbNames)
	assert.EqualValues(t, []any{"james", int64(42)}, values)
}

func TestDbValuesErrorWithoutValues(t *testing.T) {
	type input struct {
		Name string `db:"name_db"`
	}

	plan, err := AnalyzeT(input{Name: "james"}, false)
	assert.NoError(t, err)

	dbNames, values, err := plan.DbValues()
	assert.Nil(t, dbNames)
	assert.Nil(t, values)
	assert.EqualError(t, err, "Plan was analyzed without values")
}

func TestDbValuesErrorNoDbTags(t *testing.T) {
	type input struct {
		Name string `json:"name"`
	}

	plan, err := AnalyzeT(input{Name: "james"}, true)
	assert.NoError(t, err)

	dbNames, values, err := plan.DbValues()
	assert.Nil(t, dbNames)
	assert.Nil(t, values)
	assert.EqualError(t, err, "no fields have db tags")
}
