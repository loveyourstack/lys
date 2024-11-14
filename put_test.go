package lys

import (
	"context"
	"strconv"
	"testing"

	"github.com/loveyourstack/lys/internal/stores/core/coretypetest"
	"github.com/loveyourstack/lys/lysclient"
	"github.com/stretchr/testify/assert"
)

func TestPutSuccess(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// create a record with minimal values
	minInput := coretypetest.GetEmptyInput()
	newId := lysclient.MustPostToValue[coretypetest.Input, int64](t, srvApp.getRouter(), "POST", "/type-test", minInput)

	targetUrl := "/type-test/" + strconv.FormatInt(newId, 10)

	// get filled type test input
	filledInput, err := coretypetest.GetFilledInput()
	if err != nil {
		t.Fatalf("coretypetest.GetFilledInput failed: %v", err)
	}

	// PUT the filled input to the minimal record
	_ = lysclient.MustPostToValue[coretypetest.Input, string](t, srvApp.getRouter(), "PUT", targetUrl, filledInput)

	// get changed record
	filledItem := lysclient.MustDoToValue[coretypetest.Model](t, srvApp.getRouter(), "GET", targetUrl)

	// check changed record
	testFilledInput(t, filledItem.Input)
}

func TestPutFailure(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	minInput := coretypetest.GetEmptyInput()
	newId := lysclient.MustPostToValue[coretypetest.Input, int64](t, srvApp.getRouter(), "POST", "/type-test", minInput)

	targetUrl := "/type-test/" + strconv.FormatInt(newId, 10)

	// struct with unknown field
	type testS struct {
		Val string
	}
	inputTestS := testS{
		Val: "a",
	}
	_, err := lysclient.PostToValueTester[testS, string](srvApp.getRouter(), "PUT", targetUrl, inputTestS)
	assert.EqualValues(t, "unknown field: Val", err.Error(), "unknown field")

	// nil input (fails on mandatory enum val)
	_, err = lysclient.PostToValueTester[any, string](srvApp.getRouter(), "PUT", targetUrl, nil)
	assert.EqualValues(t, `invalid text: invalid input value for enum core.weekday: ""`, err.Error(), "nil")

	// empty struct (fails on mandatory enum val)
	inputTT := coretypetest.Input{}
	_, err = lysclient.PostToValueTester[coretypetest.Input, string](srvApp.getRouter(), "PUT", targetUrl, inputTT)
	assert.EqualValues(t, `invalid text: invalid input value for enum core.weekday: ""`, err.Error(), "empty struct")

	// id wrong type
	_, err = lysclient.PostToValueTester[coretypetest.Input, string](srvApp.getRouter(), "PUT", "/type-test/a", minInput)
	assert.EqualValues(t, "id not an integer", err.Error(), "id wrong type")

	// invalid id
	_, err = lysclient.PostToValueTester[coretypetest.Input, string](srvApp.getRouter(), "PUT", "/type-test/100000", minInput)
	assert.EqualValues(t, "row(s) not found", err.Error(), "invalid id")
}
