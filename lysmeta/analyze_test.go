package lysmeta

import (
	"reflect"
	"testing"
	"time"

	"github.com/loveyourstack/lys/lystype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeTSuccess(t *testing.T) {
	type Embedded struct {
		Enabled bool `db:"enabled_db" json:"enabled_json"`
	}

	type input struct {
		Name string `db:"name_db" json:"name_json"`
		Embedded
		Age int
	}

	plan, err := AnalyzeT(input{Name: "james", Embedded: Embedded{Enabled: true}, Age: 42}, false)
	require.NoError(t, err)
	assert.False(t, plan.HasValues)
	assert.EqualValues(t, 3, len(plan.Fields))

	assert.Equal(t, "Name", plan.Fields[0].Name)
	assert.Equal(t, "name_db", plan.Fields[0].DbName)
	assert.Equal(t, "name_json", plan.Fields[0].JsonKey)
	assert.Nil(t, plan.Fields[0].Value)

	assert.Equal(t, "Enabled", plan.Fields[1].Name)
	assert.Equal(t, "enabled_db", plan.Fields[1].DbName)
	assert.Equal(t, "enabled_json", plan.Fields[1].JsonKey)
	assert.Nil(t, plan.Fields[1].Value)

	assert.Equal(t, "Age", plan.Fields[2].Name)
	assert.Equal(t, "", plan.Fields[2].DbName)
	assert.Equal(t, "Age", plan.Fields[2].JsonKey)
	assert.Nil(t, plan.Fields[2].Value)
}

func TestAnalyzeTFailure(t *testing.T) {

	_, err := AnalyzeT(123, false)
	assert.EqualError(t, err, "AnalyzeT only accepts struct types, but got int", "non-struct")

	type input struct {
		age int
	}

	_, err = AnalyzeT(input{age: 18}, false)
	assert.EqualError(t, err, "struct has no exported fields", "no exported fields")
}

func TestAnalyzeAndCheckTSuccess(t *testing.T) {
	type Inner struct {
		Inner string `db:"inner_db" json:"inner_json"`
	}

	type input struct {
		Top string `db:"top_db" json:"top_json"`
		Inner
	}

	plan, err := AnalyzeAndCheckT(input{})
	require.NoError(t, err)
	assert.EqualValues(t, 2, len(plan.Fields))

	assert.Equal(t, "Top", plan.Fields[0].Name)
	assert.Equal(t, "top_db", plan.Fields[0].DbName)
	assert.Equal(t, "top_json", plan.Fields[0].JsonKey)

	assert.Equal(t, "Inner", plan.Fields[1].Name)
	assert.Equal(t, "inner_db", plan.Fields[1].DbName)
	assert.Equal(t, "inner_json", plan.Fields[1].JsonKey)

	assert.EqualValues(t, []string{"top_json", "inner_json"}, plan.JsonKeys())
}

func TestAnalyzeAndCheckTFailure(t *testing.T) {

	_, err := AnalyzeAndCheckT(123)
	assert.EqualError(t, err, "AnalyzeT failed: AnalyzeT only accepts struct types, but got int", "non-struct")

	type input struct {
		A string `db:"db2" json:"json2"`
		B string `db:"db1" json:"json1"`
		C string `db:"db2" json:"json1"`
		D string `db:"db1" json:"json2"`
	}

	_, err = AnalyzeAndCheckT(input{})
	assert.EqualError(
		t,
		err,
		"db name 'db1' is set on 2 fields, db name 'db2' is set on 2 fields, json key 'json1' is set on 2 fields, json key 'json2' is set on 2 fields",
		"duplicate tags",
	)
}

func TestGetDbName(t *testing.T) {
	assert.Equal(t, "", getDbName(""), "empty tag returns empty name")
	assert.Equal(t, "", getDbName("-"), "dash omits name")
	assert.Equal(t, "custom", getDbName("custom"), "explicit name is used")
}

func TestGetJsonKey(t *testing.T) {
	assert.Equal(t, "FieldName", getJsonKey("", "FieldName"), "empty tag falls back to field name")
	assert.Equal(t, "FieldName", getJsonKey(",omitempty", "FieldName"), "empty key with options falls back to field name")
	assert.Equal(t, "", getJsonKey("-", "FieldName"), "dash omits key")
	assert.Equal(t, "custom", getJsonKey("custom,omitempty", "FieldName"), "explicit key is used")
}

func TestGetStructFieldsSuccess(t *testing.T) {
	type Embedded struct {
		Score float64      `db:"score_db" json:"score_json"`
		DOB   lystype.Date `db:"dob_db" json:"dob_json"` // struct, but not embedded, so not recursed into
	}

	type input struct {
		ID  string  `db:"id_db" json:"id_json"`
		Ptr *string `db:"ptr_db" json:"ptr_json"`
		Embedded
		HiddenTags string `json:"-"`
	}

	dob := lystype.Date(time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC))
	ptr := lystype.ToPtr("def")

	reflVal := reflect.ValueOf(input{ID: "abc", Ptr: ptr, Embedded: Embedded{Score: 1.5, DOB: dob}, HiddenTags: "skip"})
	fields := getStructFields(reflVal, true)

	assert.EqualValues(t, 5, len(fields))

	i := 0
	assert.Equal(t, "ID", fields[i].Name)
	assert.Equal(t, "id_db", fields[i].DbName)
	assert.Equal(t, "id_json", fields[i].JsonKey)
	assert.Equal(t, "abc", fields[i].Value)

	i++
	assert.Equal(t, "Ptr", fields[i].Name)
	assert.Equal(t, "ptr_db", fields[i].DbName)
	assert.Equal(t, "ptr_json", fields[i].JsonKey)
	assert.Equal(t, ptr, fields[i].Value)

	i++
	assert.Equal(t, "Score", fields[i].Name)
	assert.Equal(t, "score_db", fields[i].DbName)
	assert.Equal(t, "score_json", fields[i].JsonKey)
	assert.Equal(t, 1.5, fields[i].Value)

	i++
	assert.Equal(t, "DOB", fields[i].Name)
	assert.Equal(t, "dob_db", fields[i].DbName)
	assert.Equal(t, "dob_json", fields[i].JsonKey)
	assert.Equal(t, "1990-05-15", fields[i].Value) // note that value is converted to string for date types

	i++
	assert.Equal(t, "HiddenTags", fields[i].Name)
	assert.Equal(t, "", fields[i].DbName)
	assert.Equal(t, "", fields[i].JsonKey)
	assert.Equal(t, "skip", fields[i].Value)
}

func TestGetStructFieldsSkipsUnexported(t *testing.T) {
	type input struct {
		Visible int `db:"visible_db" json:"visible_json"`
		hidden  int `db:"hidden_db" json:"hidden_json"`
	}

	fields := getStructFields(reflect.ValueOf(input{Visible: 7, hidden: 99}), true)
	assert.EqualValues(t, 1, len(fields))
	assert.Equal(t, "Visible", fields[0].Name)
	assert.Equal(t, 7, fields[0].Value)
}
