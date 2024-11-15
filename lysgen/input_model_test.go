package lysgen

import (
	"testing"

	"github.com/loveyourstack/lys/lyspg"
	"github.com/stretchr/testify/assert"
)

func TestGetInputTypesSuccess(t *testing.T) {

	cols := []lyspg.Column{

		// should be excluded
		{Name: "identity", DataType: "bigint", IsIdentity: true},
		{Name: "generated", DataType: "text", IsGenerated: true},
		{Name: "tracking", DataType: "text", IsTracking: true},

		// should be included
		{Name: "array", DataType: "ARRAY"},
		{Name: "bigint", DataType: "bigint"},
		{Name: "date", DataType: "date"},
		{Name: "numeric", DataType: "numeric"},
		{Name: "text", DataType: "text"},
		{Name: "time", DataType: "time"},
		{Name: "timestamp_with_time_zone", DataType: "timestamp with time zone"},
		{Name: "user_defined", DataType: "USER-DEFINED"}, // enum
	}

	actualA, err := getInput(cols, false)
	if err != nil {
		t.Fatalf("getInput failed: %s", err.Error())
	}

	expectedA := []string{
		"type Input struct {",
		"    Array  []string  `db:\"array\" json:\"array\"`",
		"    Bigint  int64  `db:\"bigint\" json:\"bigint\"`",
		"    Date  lystype.Date  `db:\"date\" json:\"date\"`",
		"    Numeric  float64  `db:\"numeric\" json:\"numeric\"`",
		"    Text  string  `db:\"text\" json:\"text\"`",
		"    Time  lystype.Time  `db:\"time\" json:\"time\"`",
		"    TimestampWithTimeZone  lystype.Datetime  `db:\"timestamp_with_time_zone\" json:\"timestamp_with_time_zone\"`",
		"    UserDefined  string  `db:\"user_defined\" json:\"user_defined\"`",
		"}",
	}

	assert.EqualValues(t, expectedA, actualA)
}

func TestGetInputTypesFailure(t *testing.T) {

	cols := []lyspg.Column{
		{Name: "unknown", DataType: "unknown"},
	}

	_, err := getInput(cols, false)
	assert.EqualError(t, err, "GetGoDataTypeFromPg failed: no go type found for pgType: unknown")
}

func TestGetInputWithValidationSuccess(t *testing.T) {

	cols := []lyspg.Column{
		{Name: "text", DataType: "text"},
	}

	actualA, err := getInput(cols, true)
	if err != nil {
		t.Fatalf("getInput failed: %s", err.Error())
	}

	expectedA := []string{
		"type Input struct {",
		"    Text  string  `db:\"text\" json:\"text\" validate:\"required\"`",
		"}",
	}

	assert.EqualValues(t, expectedA, actualA)
}

func TestGetModelSuccess(t *testing.T) {

	cols := []lyspg.Column{

		// should be excluded
		{Name: "text", DataType: "text"},

		// should be included
		{Name: "id", DataType: "bigint", IsIdentity: true},
		{Name: "generated", DataType: "text", IsGenerated: true},
		{Name: "entry_at", DataType: "timestamp with time zone", IsTracking: true},
	}

	parentCols := []lyspg.Column{
		{SchemaName: "core", TableName: "parent_lvl1", Name: "name", DataType: "text"}, // lvl1 = directly above in hierarchy
	}

	childFks := []lyspg.ForeignKey{
		{ChildTable: "child_lvl1"},
	}

	actualA, err := getModel(cols, parentCols, childFks)
	if err != nil {
		t.Fatalf("getModel failed: %s", err.Error())
	}

	expectedA := []string{
		"type Model struct {",
		"    Id  int64  `db:\"id\" json:\"id\"`",
		"    Generated  string  `db:\"generated\" json:\"generated\"`",
		"    EntryAt  lystype.Datetime  `db:\"entry_at\" json:\"entry_at\"`",
		"    ParentLvl1Name  string  `db:\"parent_lvl1_name\" json:\"parent_lvl1_name\"`",
		"    ChildLvl1Count  int  `db:\"child_lvl1_count\" json:\"child_lvl1_count\"`",
		"    Input",
		"}",
	}

	assert.EqualValues(t, expectedA, actualA)
}
