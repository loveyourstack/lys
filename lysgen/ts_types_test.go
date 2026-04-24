package lysgen

import (
	"testing"

	"github.com/loveyourstack/lys/lyspg"
	"github.com/stretchr/testify/assert"
)

func TestGetTsInputIFaceSuccess(t *testing.T) {

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

	actualA, err := getTsInputIFace("MyTable", cols)
	if err != nil {
		t.Fatalf("getTsInputIFace failed: %s", err.Error())
	}

	expectedA := []string{
		"export interface MyTableInput {",
		"  array: string[] | undefined",
		"  bigint: number | undefined",
		"  date: string | undefined",
		"  numeric: number | undefined",
		"  text: string | undefined",
		"  time: string | undefined",
		"  timestamp_with_time_zone: Date | undefined",
		"  user_defined: string | undefined",
		"}",
	}

	assert.EqualValues(t, expectedA, actualA)
}

func TestGetTsInputIFaceFailure(t *testing.T) {

	cols := []lyspg.Column{
		{Name: "unknown", DataType: "unknown"},
	}

	_, err := getTsInputIFace("MyTable", cols)
	assert.EqualError(t, err, "GetTsDataTypeFromPg failed: no Typescript type found for pgType: unknown")
}

func TestGetTsIFaceSuccess(t *testing.T) {

	cols := []lyspg.Column{

		// should be excluded
		{Name: "text", DataType: "text"},

		// should be included
		{Name: "id", DataType: "bigint", IsIdentity: true},
		{Name: "generated", DataType: "text", IsGenerated: true},
		{Name: "created_at", DataType: "timestamp with time zone", IsTracking: true},
	}

	parentCols := []lyspg.Column{
		{SchemaName: "core", TableName: "parent_lvl1", Name: "name", DataType: "text"}, // lvl1 = directly above in hierarchy
	}

	childFks := []lyspg.ForeignKey{
		{ChildTable: "child_lvl1"},
	}

	actualA, err := getTsIFace("MyTable", cols, parentCols, childFks)
	if err != nil {
		t.Fatalf("getTsIFace failed: %s", err.Error())
	}

	expectedA := []string{
		"export interface MyTable extends MyTableInput {",
		"  id: number",
		"  generated: string",
		"  created_at: Date",
		"  parent_lvl1_name: string",
		"  child_lvl1_count: number",
		"}",
	}

	assert.EqualValues(t, expectedA, actualA)
}
