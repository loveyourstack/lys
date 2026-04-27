package lysmeta

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyzeAndCheckTSuccess(t *testing.T) {
	type Inner struct {
		Inner string `db:"inner_db" json:"inner_json"`
	}

	type input struct {
		Top string `db:"top_db" json:"top_json"`
		Inner
	}

	plan, err := AnalyzeAndCheckT(input{})
	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(plan.Fields))

	assert.Equal(t, "Top", plan.Fields[0].Name)
	assert.Equal(t, "top_db", plan.Fields[0].DbName)
	assert.Equal(t, "top_json", plan.Fields[0].JsonKey)

	assert.Equal(t, "Inner", plan.Fields[1].Name)
	assert.Equal(t, "inner_db", plan.Fields[1].DbName)
	assert.Equal(t, "inner_json", plan.Fields[1].JsonKey)

	assert.EqualValues(t, []string{"top_json", "inner_json"}, plan.JsonKeys())
}

func TestAnalyzeAndCheckTFailureDuplicatesDeterministic(t *testing.T) {
	type input struct {
		A string `db:"db2" json:"json2"`
		B string `db:"db1" json:"json1"`
		C string `db:"db2" json:"json1"`
		D string `db:"db1" json:"json2"`
	}

	_, err := AnalyzeAndCheckT(input{})
	assert.EqualError(
		t,
		err,
		"db name 'db1' is set on 2 fields, db name 'db2' is set on 2 fields, json key 'json1' is set on 2 fields, json key 'json2' is set on 2 fields",
	)
}

func TestAnalyzeTFailureNonStruct(t *testing.T) {
	_, err := AnalyzeT(123, false)
	assert.EqualError(t, err, "AnalyzeT only accepts struct types, but got int")
}

func TestAnalyzeTGetValuesAndJsonKeyFallback(t *testing.T) {
	type input struct {
		Name  string `db:"name_db" json:"name_json"`
		Alias string `db:"alias_db" json:",omitempty"`
	}

	plan, err := AnalyzeT(input{Name: "james", Alias: "jam"}, true)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(plan.Fields))

	assert.Equal(t, "Name", plan.Fields[0].Name)
	assert.Equal(t, "name_json", plan.Fields[0].JsonKey)
	assert.Equal(t, "james", plan.Fields[0].Value)

	assert.Equal(t, "Alias", plan.Fields[1].Name)
	assert.Equal(t, "Alias", plan.Fields[1].JsonKey)
	assert.Equal(t, "jam", plan.Fields[1].Value)
}

func TestGetStructFields2SkipsUnexported(t *testing.T) {
	type input struct {
		Visible int `db:"visible_db" json:"visible_json"`
		hidden  int `db:"hidden_db" json:"hidden_json"`
	}

	fields := getStructFields(reflect.ValueOf(input{Visible: 7, hidden: 99}), true)
	assert.EqualValues(t, 1, len(fields))
	assert.Equal(t, "Visible", fields[0].Name)
	assert.Equal(t, 7, fields[0].Value)
}
