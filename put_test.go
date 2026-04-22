package lys

import (
	"context"
	"strconv"
	"testing"

	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/loveyourstack/lys/lysclient"
	"github.com/stretchr/testify/assert"
)

func TestPutSuccess(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	// create a record with minimal values
	minInput := coretypetestm.GetEmptyInput()
	newId := lysclient.MustPostToValue[coretypetestm.Input, int64](ctx, t, srvApp.getRouter(), "POST", "/type-test", minInput)

	targetUrl := "/type-test/" + strconv.FormatInt(newId, 10)

	// get filled type test input
	filledInput, err := coretypetestm.GetFilledInput()
	if err != nil {
		t.Fatalf("coretypetestm.GetFilledInput failed: %v", err)
	}

	// PUT the filled input to the minimal record
	_ = lysclient.MustPostToValue[coretypetestm.Input, string](ctx, t, srvApp.getRouter(), "PUT", targetUrl, filledInput)

	// get changed record
	filledItem := lysclient.MustDoToValue[coretypetestm.Model](ctx, t, srvApp.getRouter(), "GET", targetUrl)

	// check changed record
	coretypetestm.TestFilledInput(t, filledItem.Input)
}

func TestPutFailure(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	minInput := coretypetestm.GetEmptyInput()
	newId := lysclient.MustPostToValue[coretypetestm.Input, int64](ctx, t, srvApp.getRouter(), "POST", "/type-test", minInput)

	targetUrl := "/type-test/" + strconv.FormatInt(newId, 10)

	// struct with unknown field
	type testS struct {
		Val string
	}
	inputTestS := testS{
		Val: "a",
	}
	_, err := lysclient.PostToValueTester[testS, string](ctx, srvApp.getRouter(), "PUT", targetUrl, inputTestS)
	assert.EqualValues(t, "unknown field: Val", err.Error(), "unknown field")

	// nil input (fails on mandatory enum val)
	_, err = lysclient.PostToValueTester[any, string](ctx, srvApp.getRouter(), "PUT", targetUrl, nil)
	assert.EqualValues(t, `invalid text: invalid input value for enum core.weekday: ""`, err.Error(), "nil")

	// empty struct (fails on mandatory enum val)
	inputTT := coretypetestm.Input{}
	_, err = lysclient.PostToValueTester[coretypetestm.Input, string](ctx, srvApp.getRouter(), "PUT", targetUrl, inputTT)
	assert.EqualValues(t, `invalid text: invalid input value for enum core.weekday: ""`, err.Error(), "empty struct")

	// id wrong type
	_, err = lysclient.PostToValueTester[coretypetestm.Input, string](ctx, srvApp.getRouter(), "PUT", "/type-test/a", minInput)
	assert.EqualValues(t, "id not an integer", err.Error(), "id wrong type")

	// invalid id
	_, err = lysclient.PostToValueTester[coretypetestm.Input, string](ctx, srvApp.getRouter(), "PUT", "/type-test/100000", minInput)
	assert.EqualValues(t, "row(s) not found", err.Error(), "invalid id")
}
