package lys

import (
	"context"
	"strconv"
	"testing"

	"github.com/loveyourstack/lys/internal/stores/core/coretypetest"
	"github.com/loveyourstack/lys/lysclient"
	"github.com/stretchr/testify/assert"
)

func TestDeleteSuccess(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// create record
	minInput := coretypetest.GetEmptyInput()
	minItem := lysclient.MustPostToValue[coretypetest.Input, coretypetest.Model](t, srvApp.getRouter(), "POST", "/type-test", minInput)

	targetUrl := "/type-test/" + strconv.FormatInt(minItem.Id, 10)

	// delete it
	_ = lysclient.MustDoToValue[string](t, srvApp.getRouter(), "DELETE", targetUrl)

	// try to select it
	_, err := lysclient.DoToValueTester[string](srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, "row(s) not found", err.Error())
}

func TestDeleteFailure(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// id wrong type
	_, err := lysclient.DoToValueTester[string](srvApp.getRouter(), "DELETE", "/type-test/a")
	assert.EqualValues(t, "id not an integer", err.Error(), "id wrong type")

	// invalid id
	_, err = lysclient.DoToValueTester[string](srvApp.getRouter(), "DELETE", "/type-test/100000")
	assert.EqualValues(t, "row(s) not found", err.Error(), "invalid id")
}
