package lys

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/loveyourstack/lys/lysclient"
	"github.com/stretchr/testify/assert"
)

func TestImportSuccess(t *testing.T) {

	schemaName := "core"
	tableName := "import_test"
	pkColName := "id"

	ctx := context.Background()

	srvApp := mustGetSrvApp(t, ctx)
	defer srvApp.Db.Close()

	// with empty inputs
	inputs := []coretypetestm.Input{}
	for range 10 {
		input := coretypetestm.GetEmptyInput()
		inputs = append(inputs, input)
	}

	rowsAffected := lysclient.MustPostToValue[[]coretypetestm.Input, int64](t, srvApp.getRouter(), "POST", "/import-test/import", inputs)
	assert.EqualValues(t, 10, rowsAffected, "type test - empty")

	// test last inserted value
	stmt := fmt.Sprintf("SELECT * FROM %s.%s ORDER BY %s DESC LIMIT 1", schemaName, tableName, pkColName)
	rows, _ := srvApp.Db.Query(ctx, stmt)
	item, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[coretypetestm.Model])
	if err != nil {
		t.Fatalf("pgx.CollectExactlyOneRow failed: %v", err)
	}
	coretypetestm.TestEmptyInput(t, item.Input)

	// with filled inputs
	inputs = []coretypetestm.Input{}
	for range 10 {
		input, err := coretypetestm.GetFilledInput()
		if err != nil {
			t.Fatalf("coretypetestm.GetFilledInput failed: %v", err)
		}
		inputs = append(inputs, input)
	}

	rowsAffected = lysclient.MustPostToValue[[]coretypetestm.Input, int64](t, srvApp.getRouter(), "POST", "/import-test/import", inputs)
	assert.EqualValues(t, 10, rowsAffected, "type test - filled")

	// test last inserted value
	stmt = fmt.Sprintf("SELECT * FROM %s.%s ORDER BY %s DESC LIMIT 1", schemaName, tableName, pkColName)
	rows, _ = srvApp.Db.Query(ctx, stmt)
	item, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[coretypetestm.Model])
	if err != nil {
		t.Fatalf("pgx.CollectExactlyOneRow failed: %v", err)
	}
	coretypetestm.TestFilledInput(t, item.Input)
}

func TestImportParamsFailure(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// struct with unknown field
	type testS struct {
		Val string
	}
	inputTestS := testS{
		Val: "a",
	}
	testSA := []testS{}
	testSA = append(testSA, inputTestS)
	testSA = append(testSA, inputTestS)

	_, err := lysclient.PostArrayToValueTester[testS, int64](srvApp.getRouter(), "POST", "/import-test/import", testSA)
	assert.EqualValues(t, "unknown field: Val", err.Error(), "unknown field")

	// no inputs (nil)
	_, err = lysclient.PostArrayToValueTester[any, int64](srvApp.getRouter(), "POST", "/import-test/import", nil)
	assert.EqualValues(t, "no inputs found", err.Error(), "nil")

	// no inputs (empty slice)
	inputTTA := []coretypetestm.Input{}
	_, err = lysclient.PostArrayToValueTester[coretypetestm.Input, int64](srvApp.getRouter(), "POST", "/import-test/import", inputTTA)
	assert.EqualValues(t, "no inputs found", err.Error(), "empty slice")
}

func TestImportValidationFailure(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// an input fails validation
	inputs := []coretypetestm.Input{}
	input := coretypetestm.GetEmptyInput()
	inputs = append(inputs, input)
	input.CText = "fail" // see coreimporttest store Validate()
	inputs = append(inputs, input)

	_, err := lysclient.PostArrayToValueTester[coretypetestm.Input, int64](srvApp.getRouter(), "POST", "/import-test/import", inputs)
	assert.EqualValues(t, "line 2: CText is invalid", err.Error(), "")
}
