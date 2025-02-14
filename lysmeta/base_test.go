package lysmeta

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mustAnalyzeStructs(t *testing.T, reflVals ...reflect.Value) Result {
	ret, err := AnalyzeStructs(reflVals...)
	if err != nil {
		t.Fatalf("AnalyzeStructs failed: %v", err)
	}
	return ret
}

func TestAnalyzeStructsSuccess(t *testing.T) {

	type inner struct {
		A string `db:"a1" json:"a2,omitempty"` // omitempty should be removed
		C string `db:"-" json:"-"`             // should be ignored
		D string `db:"" json:""`               // should be ignored
	}

	// single struct
	res := mustAnalyzeStructs(t, reflect.ValueOf(&inner{}).Elem())
	assert.EqualValues(t, []string{"a1"}, res.DbTags, "single: DbTags")
	assert.EqualValues(t, []string{"a2"}, res.JsonTags, "single: JsonTags")
	jttMap := make(map[string]string)
	jttMap["a2"] = "string"
	assert.EqualValues(t, jttMap, res.JsonTagTypeMap, "single: JsonTagTypeMap")

	type outer struct {
		B int64 `db:"b1" json:"b2"`
		inner
	}

	// multiple structs with embedding: embedded gets ignored
	res = mustAnalyzeStructs(t, reflect.ValueOf(&outer{}).Elem(), reflect.ValueOf(&inner{}).Elem())
	assert.EqualValues(t, []string{"b1", "a1"}, res.DbTags, "multiple: DbTags")
	assert.EqualValues(t, []string{"b2", "a2"}, res.JsonTags, "multiple: JsonTags")
	jttMap = make(map[string]string)
	jttMap["b2"] = "int64"
	jttMap["a2"] = "string"
	assert.EqualValues(t, jttMap, res.JsonTagTypeMap, "multiple: JsonTagTypeMap")

	type noTags struct {
		A string
	}

	// no tags
	res = mustAnalyzeStructs(t, reflect.ValueOf(&noTags{}).Elem())
	assert.EqualValues(t, 0, len(res.DbTags), "no tags: DbTags")
	assert.EqualValues(t, 0, len(res.JsonTags), "no tags: JsonTags")
	assert.EqualValues(t, 0, len(res.JsonTagTypeMap), "no tags: JsonTagTypeMap")
}

func TestAnalyzeStructsFailure(t *testing.T) {

	// duplicated db tags in single struct
	type dbDup struct {
		A string `db:"a1" json:"a2,omitempty"`
		B string `db:"a1" json:"b2"`
	}

	_, err := AnalyzeStructs(reflect.ValueOf(&dbDup{}).Elem())
	assert.EqualError(t, err, "db tag 'a1' is set on 2 fields", "dup db tags: single")

	// duplicated db tags over multiple structs
	type dbDupA struct {
		A string `db:"a1" json:"a2"`
	}
	type dbDupB struct {
		B string `db:"a1" json:"b2"`
	}

	_, err = AnalyzeStructs(reflect.ValueOf(&dbDupA{}).Elem(), reflect.ValueOf(&dbDupB{}).Elem())
	assert.EqualError(t, err, "db tag 'a1' is set on 2 fields", "dup db tags: multiple")

	// duplicated json tags in single struct
	type jsonDup struct {
		A string `db:"a1" json:"a2,omitempty"`
		B string `db:"b1" json:"a2"`
	}

	_, err = AnalyzeStructs(reflect.ValueOf(&jsonDup{}).Elem())
	assert.EqualError(t, err, "json tag 'a2' is set on 2 fields", "dup json tags: single")

	// duplicated json tags over multiple structs
	type jsonDupA struct {
		A string `db:"a1" json:"a2"`
	}
	type jsonDupB struct {
		B string `db:"b1" json:"a2"`
	}

	_, err = AnalyzeStructs(reflect.ValueOf(&jsonDupA{}).Elem(), reflect.ValueOf(&jsonDupB{}).Elem())
	assert.EqualError(t, err, "json tag 'a2' is set on 2 fields", "dup json tags: multiple")
}
