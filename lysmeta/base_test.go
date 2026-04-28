package lysmeta

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDbFuncs(t *testing.T) {
	type input struct {
		Name string `db:"name_db" json:"name"`
		Age  int64  `db:"age_db" json:"age"`
		Skip string `json:"skip"`
	}

	plan, err := AnalyzeT(input{Name: "james", Age: 42, Skip: "ignored"}, true)
	require.NoError(t, err)

	t.Run("DbNames", func(t *testing.T) {
		dbNames := plan.DbNames()
		assert.EqualValues(t, []string{"name_db", "age_db"}, dbNames)
	})

	t.Run("DbValues", func(t *testing.T) {
		dbNames, values, err := plan.DbValues()
		assert.NoError(t, err)
		assert.EqualValues(t, []string{"name_db", "age_db"}, dbNames)
		assert.EqualValues(t, []any{"james", int64(42)}, values)
	})
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
	assert.EqualError(t, err, "Plan was created without values")
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

func TestJsonFuncs(t *testing.T) {
	plan := Plan{
		Fields: []Field{
			{Name: "Name", Type: reflect.TypeFor[string](), JsonKey: "name", DbName: "name_db"},
			{Name: "Age", Type: reflect.TypeFor[int64](), JsonKey: "age", DbName: "age_db"},
			{Name: "Hidden", Type: reflect.TypeFor[bool](), JsonKey: "", DbName: "hidden_db"},
			{Name: "Score", Type: reflect.TypeFor[float64](), JsonKey: "score", DbName: ""},
		},
	}

	t.Run("JsonKeys", func(t *testing.T) {
		k := plan.JsonKeys()
		assert.Equal(t, []string{"name", "age", "score"}, k)
	})

	t.Run("JsonKeyTypeMap", func(t *testing.T) {
		m := plan.JsonKeyTypeMap()
		assert.EqualValues(t, 3, len(m))
		assert.Equal(t, reflect.TypeFor[string](), m["name"])
		assert.Equal(t, reflect.TypeFor[int64](), m["age"])
		assert.Equal(t, reflect.TypeFor[float64](), m["score"])
		_, exists := m[""]
		assert.False(t, exists)
	})

	t.Run("JsonKeyDbNameMap", func(t *testing.T) {
		m := plan.JsonKeyDbNameMap()
		assert.EqualValues(t, 2, len(m))
		assert.Equal(t, "name_db", m["name"])
		assert.Equal(t, "age_db", m["age"])
		_, exists := m["score"]
		assert.False(t, exists)
		_, exists = m[""]
		assert.False(t, exists)
	})
}
