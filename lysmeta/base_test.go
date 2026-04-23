package lysmeta

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mustAnalyzeStruct(t *testing.T, reflVal reflect.Value) Result {
	ret, err := AnalyzeStruct(reflVal)
	if err != nil {
		t.Fatalf("AnalyzeStruct failed: %v", err)
	}
	return ret
}

func TestAnalyzeStructSuccess(t *testing.T) {

	type inner struct {
		A string `db:"a1" json:"a2,omitempty"` // omitempty should be removed
		C string `db:"-" json:"-"`             // should be ignored
		D string `db:"" json:""`               // should be ignored
	}

	// regular struct
	res := mustAnalyzeStruct(t, reflect.ValueOf(&inner{}).Elem())
	assert.EqualValues(t, []string{"a1"}, res.DbTags, "regular: DbTags")
	assert.EqualValues(t, []string{"a2"}, res.JsonTags, "regular: JsonTags")

	jttMap := make(map[string]reflect.Type)
	jttMap["a2"] = reflect.TypeFor[string]()
	assert.EqualValues(t, jttMap, res.JsonTagTypeMap, "regular: JsonTagTypeMap")

	type outer struct {
		B int64 `db:"b1" json:"b2"`
		inner
	}

	// struct with embedding: embedded is included and tags are combined
	res = mustAnalyzeStruct(t, reflect.ValueOf(&outer{}).Elem())
	assert.EqualValues(t, []string{"b1", "a1"}, res.DbTags, "with embedding: DbTags")
	assert.EqualValues(t, []string{"b2", "a2"}, res.JsonTags, "with embedding: JsonTags")

	jttMap = make(map[string]reflect.Type)
	jttMap["b2"] = reflect.TypeFor[int64]()
	jttMap["a2"] = reflect.TypeFor[string]()
	assert.EqualValues(t, jttMap, res.JsonTagTypeMap, "with embedding: JsonTagTypeMap")

	type noTags struct {
		A string
	}

	// no tags
	res = mustAnalyzeStruct(t, reflect.ValueOf(&noTags{}).Elem())
	assert.EqualValues(t, 0, len(res.DbTags), "no tags: DbTags")
	assert.EqualValues(t, 0, len(res.JsonTags), "no tags: JsonTags")
	assert.EqualValues(t, 0, len(res.JsonTagTypeMap), "no tags: JsonTagTypeMap")

	// embedding with no tags: embedded is skipped and tags are from outer only
	type outer2 struct {
		B int64 `db:"b1" json:"b2"`
		noTags
	}

	// embedded struct has no db or json tags: it is skipped
	res = mustAnalyzeStruct(t, reflect.ValueOf(&outer2{}).Elem())
	assert.EqualValues(t, []string{"b1"}, res.DbTags, "embedding with no tags: DbTags")
	assert.EqualValues(t, []string{"b2"}, res.JsonTags, "embedding with no tags: JsonTags")

	jttMap = make(map[string]reflect.Type)
	jttMap["b2"] = reflect.TypeFor[int64]()
	assert.EqualValues(t, jttMap, res.JsonTagTypeMap, "embedding with no tags: JsonTagTypeMap")
}

func TestAnalyzeStructFailure(t *testing.T) {

	// duplicated db tags
	type dbDup struct {
		A string `db:"a1" json:"a2,omitempty"`
		B string `db:"a1" json:"b2"`
	}

	_, err := AnalyzeStruct(reflect.ValueOf(&dbDup{}).Elem())
	assert.EqualError(t, err, "db tag 'a1' is set on 2 fields", "dup db tags")

	// duplicated db tags via embedding
	type dbDupInner struct {
		A string `db:"a1" json:"a2,omitempty"`
	}
	type dbDupOuter struct {
		B string `db:"a1" json:"b2"`
		dbDupInner
	}

	_, err = AnalyzeStruct(reflect.ValueOf(&dbDupOuter{}).Elem())
	assert.EqualError(t, err, "db tag 'a1' is set on 2 fields", "dup db tags: via embedded")

	// duplicated json tags
	type jsonDup struct {
		A string `db:"a1" json:"a2,omitempty"`
		B string `db:"b1" json:"a2"`
	}

	_, err = AnalyzeStruct(reflect.ValueOf(&jsonDup{}).Elem())
	assert.EqualError(t, err, "json tag 'a2' is set on 2 fields", "dup json tags")

	// duplicated json tags via embedding
	type jsonDupInner struct {
		A string `db:"a1" json:"a2,omitempty"`
	}
	type jsonDupOuter struct {
		B string `db:"b1" json:"a2"`
		jsonDupInner
	}

	_, err = AnalyzeStruct(reflect.ValueOf(&jsonDupOuter{}).Elem())
	assert.EqualError(t, err, "json tag 'a2' is set on 2 fields", "dup json tags: via embedded")
}
