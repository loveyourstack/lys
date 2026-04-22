package lys

import (
	"context"
	"strconv"
	"testing"

	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/loveyourstack/lys/lysclient"
	"github.com/stretchr/testify/assert"
)

func TestDeleteSuccess(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	// create record
	minInput := coretypetestm.GetEmptyInput()
	newId := lysclient.MustPostToValue[coretypetestm.Input, int64](ctx, t, srvApp.getRouter(), "POST", "/type-test", minInput)

	targetUrl := "/type-test/" + strconv.FormatInt(newId, 10)

	// delete it
	_ = lysclient.MustDoToValue[string](ctx, t, srvApp.getRouter(), "DELETE", targetUrl)

	// try to select it
	_, err := lysclient.DoToValueTester[string](ctx, srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, "row(s) not found", err.Error())
}

func TestDeleteFailure(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	// id wrong type
	_, err := lysclient.DoToValueTester[string](ctx, srvApp.getRouter(), "DELETE", "/type-test/a")
	assert.EqualValues(t, "id not an integer", err.Error(), "id wrong type")

	// invalid id
	_, err = lysclient.DoToValueTester[string](ctx, srvApp.getRouter(), "DELETE", "/type-test/100000")
	assert.EqualValues(t, "row(s) not found", err.Error(), "invalid id")
}
