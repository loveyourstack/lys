package lys

import (
	"context"
	"strconv"
	"testing"

	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/loveyourstack/lys/lysclient"
	"github.com/stretchr/testify/assert"
)

func TestPatchSuccess(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// create a record with minimal values
	minInput := coretypetestm.GetEmptyInput()
	newId := lysclient.MustPostToValue[coretypetestm.Input, int64](t, srvApp.getRouter(), "POST", "/type-test", minInput)

	targetUrl := "/type-test/" + strconv.FormatInt(newId, 10)

	// get filled type test input
	filledInput, err := coretypetestm.GetFilledInput()
	if err != nil {
		t.Fatalf("coretypetestm.GetFilledInput failed: %v", err)
	}

	// PATCH the filled input to the minimal record
	_ = lysclient.MustPostToValue[coretypetestm.Input, string](t, srvApp.getRouter(), "PATCH", targetUrl, filledInput)
	//fmt.Printf("%+v", item)

	// get changed record
	filledItem := lysclient.MustDoToValue[coretypetestm.Model](t, srvApp.getRouter(), "GET", targetUrl)

	// check changed record
	coretypetestm.TestFilledInput(t, filledItem.Input)
}

func TestPatchFailure(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	minInput := coretypetestm.GetEmptyInput()
	newId := lysclient.MustPostToValue[coretypetestm.Input, int64](t, srvApp.getRouter(), "POST", "/type-test", minInput)

	targetUrl := "/type-test/" + strconv.FormatInt(newId, 10)

	// struct with unknown field
	type testS struct {
		Val string
	}
	inputTestS := testS{
		Val: "a",
	}
	_, err := lysclient.PostToValueTester[testS, string](srvApp.getRouter(), "PATCH", targetUrl, inputTestS)
	assert.EqualValues(t, "invalid field: Val", err.Error(), "invalid field")

	// nil input
	_, err = lysclient.PostToValueTester[any, string](srvApp.getRouter(), "PATCH", targetUrl, nil)
	assert.EqualValues(t, "no assignments found", err.Error(), "nil")

	// empty struct
	inputTT := coretypetestm.Input{}
	_, err = lysclient.PostToValueTester[coretypetestm.Input, string](srvApp.getRouter(), "PATCH", targetUrl, inputTT)
	assert.EqualValues(t, "invalid text: invalid input value for enum core.weekday: \"\"", err.Error(), "empty struct")

	// id wrong type
	_, err = lysclient.PostToValueTester[coretypetestm.Input, string](srvApp.getRouter(), "PATCH", "/type-test/a", minInput)
	assert.EqualValues(t, "id not an integer", err.Error(), "id wrong type")

	// invalid id
	_, err = lysclient.PostToValueTester[coretypetestm.Input, string](srvApp.getRouter(), "PATCH", "/type-test/100000", minInput)
	assert.EqualValues(t, "row(s) not found", err.Error(), "invalid id")
}
