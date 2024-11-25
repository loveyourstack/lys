package lys

import (
	"context"
	"testing"

	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/loveyourstack/lys/lysclient"
	"github.com/stretchr/testify/assert"
)

func TestGetByIdSuccess(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	targetUrl := "/type-test/1"
	item := lysclient.MustDoToValue[coretypetestm.Model](t, srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, true, item.CBool)
}

func TestGetByIdFailure(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// id wrong type
	targetUrl := "/type-test/a"
	_, err := lysclient.DoToValueTester[coretypetestm.Model](srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, "id not an integer", err.Error())

	// id doesn't exist
	targetUrl = "/type-test/100000"
	_, err = lysclient.DoToValueTester[coretypetestm.Model](srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, "row(s) not found", err.Error())
}
