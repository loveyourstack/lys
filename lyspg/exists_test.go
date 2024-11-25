package lyspg

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func mustExists(t *testing.T, db *pgxpool.Pool, schemaName, tableName, colName string, val any) bool {
	ret, err := Exists(context.Background(), db, schemaName, tableName, colName, val)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	return ret
}

func mustExistsConditions(t *testing.T, db *pgxpool.Pool, schemaName, tableName, match string, colValMap map[string]any) bool {
	ret, err := ExistsConditions(context.Background(), db, schemaName, tableName, match, colValMap)
	if err != nil {
		t.Fatalf("ExistsConditions failed: %v", err)
	}
	return ret
}

func TestExistsSuccess(t *testing.T) {

	db := mustGetDb(t, context.Background())
	defer db.Close()

	// int, true
	ret := mustExists(t, db, "core", "exists_test", "c_int", 1)
	assert.EqualValues(t, true, ret, "c_int, true")

	// int, false
	ret = mustExists(t, db, "core", "exists_test", "c_int", 100000)
	assert.EqualValues(t, false, ret, "c_int, false")

	// text, true
	ret = mustExists(t, db, "core", "exists_test", "c_text", "a")
	assert.EqualValues(t, true, ret, "c_text, true")

	// text, false
	ret = mustExists(t, db, "core", "exists_test", "c_text", "aaaaaa")
	assert.EqualValues(t, false, ret, "c_text, false")

	// text null, true
	ret = mustExists(t, db, "core", "exists_test", "c_text", nil)
	assert.EqualValues(t, true, ret, "c_text null, true")

	// int null, false
	ret = mustExists(t, db, "core", "exists_test", "c_int", nil)
	assert.EqualValues(t, false, ret, "c_int null, false")
}

func TestExistsConditionsSuccess(t *testing.T) {

	db := mustGetDb(t, context.Background())
	defer db.Close()

	// AND, true
	colValMap := make(map[string]any)
	colValMap["c_int"] = 1
	colValMap["c_text"] = "a"
	ret := mustExistsConditions(t, db, "core", "exists_test", "AND", colValMap)
	assert.EqualValues(t, true, ret, "AND, true")

	// AND with NULL, true
	colValMap = make(map[string]any)
	colValMap["c_int"] = 3
	colValMap["c_text"] = nil
	ret = mustExistsConditions(t, db, "core", "exists_test", "AND", colValMap)
	assert.EqualValues(t, true, ret, "AND with NULL, true")

	// AND, false
	colValMap = make(map[string]any)
	colValMap["c_int"] = 1000000
	colValMap["c_text"] = "a"
	ret = mustExistsConditions(t, db, "core", "exists_test", "AND", colValMap)
	assert.EqualValues(t, false, ret, "AND, false")

	// AND with NULL, false
	colValMap = make(map[string]any)
	colValMap["c_int"] = nil
	colValMap["c_text"] = "a"
	ret = mustExistsConditions(t, db, "core", "exists_test", "AND", colValMap)
	assert.EqualValues(t, false, ret, "AND with NULL, false")

	// OR, true
	colValMap = make(map[string]any)
	colValMap["c_int"] = 1000000
	colValMap["c_text"] = "a"
	ret = mustExistsConditions(t, db, "core", "exists_test", "OR", colValMap)
	assert.EqualValues(t, true, ret, "OR, true")

	// OR with NULL, true
	colValMap = make(map[string]any)
	colValMap["c_int"] = 1000000
	colValMap["c_text"] = nil
	ret = mustExistsConditions(t, db, "core", "exists_test", "OR", colValMap)
	assert.EqualValues(t, true, ret, "OR with NULL, true")

	// OR, false
	colValMap = make(map[string]any)
	colValMap["c_int"] = 1000000
	colValMap["c_text"] = "aaaaaaa"
	ret = mustExistsConditions(t, db, "core", "exists_test", "OR", colValMap)
	assert.EqualValues(t, false, ret, "OR, false")

	// OR with NULL, false
	colValMap = make(map[string]any)
	colValMap["c_int"] = nil
	colValMap["c_text"] = "aaaaaaa"
	ret = mustExistsConditions(t, db, "core", "exists_test", "OR", colValMap)
	assert.EqualValues(t, false, ret, "OR with NULL, false")
}

func TestExistsConditionsFailure(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(t, ctx)
	defer db.Close()

	// invalid match string
	colValMap := make(map[string]any)
	colValMap["c_int"] = 1
	colValMap["c_text"] = "a"
	_, err := ExistsConditions(ctx, db, "core", "exists_test", "xxx", colValMap)
	assert.EqualError(t, err, "match must be 'AND' or 'OR'", "invalid match")
}
