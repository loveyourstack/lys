package lys

import (
	"context"
	"testing"

	"github.com/loveyourstack/lys/internal/stores/core/corearchivetest"
	"github.com/loveyourstack/lys/lysclient"
	"github.com/stretchr/testify/assert"
)

func TestArchiveSuccess(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	baseUrl := "/archive-test/1"

	// get id 1
	item := lysclient.MustDoToValue[corearchivetest.Model](t, srvApp.getRouter(), "GET", baseUrl)
	assert.EqualValues(t, 1, *item.CInt, "before sd")
	assert.Nil(t, item.CText, "before sd")

	// archive id 1
	_ = lysclient.MustDoToValue[string](t, srvApp.getRouter(), "DELETE", baseUrl+"/archive")

	// try to get id 1
	_, err := lysclient.DoToValueTester[string](srvApp.getRouter(), "GET", baseUrl)
	assert.EqualValues(t, "invalid id", err.Error())

	// restore id 1
	_ = lysclient.MustDoToValue[string](t, srvApp.getRouter(), "POST", baseUrl+"/restore")

	// get id 1
	item = lysclient.MustDoToValue[corearchivetest.Model](t, srvApp.getRouter(), "GET", baseUrl)
	assert.EqualValues(t, 1, *item.CInt, "after sd")
	assert.Nil(t, item.CText, "after sd")
}

func TestArchiveFailure(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	baseUrl := "/archive-test/100000"

	// archive invalid id
	_, err := lysclient.DoToValueTester[string](srvApp.getRouter(), "DELETE", baseUrl+"/archive")
	assert.EqualValues(t, "invalid id", err.Error(), "invalid id")

	// restore invalid id
	_, err = lysclient.DoToValueTester[string](srvApp.getRouter(), "POST", baseUrl+"/restore")
	assert.EqualValues(t, "invalid id", err.Error(), "invalid id")
}
