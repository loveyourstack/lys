package lys

import (
	"context"
	"testing"

	"github.com/loveyourstack/lys/internal/stores/core/corearchivetestm"
	"github.com/loveyourstack/lys/lysclient"
	"github.com/stretchr/testify/assert"
)

func TestArchiveByIdSuccess(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	targetUrl := "/archive-test/1"

	// get id 1
	item := lysclient.MustDoToValue[corearchivetestm.Model](t, srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, 1, *item.CInt, "before archive")
	assert.Nil(t, item.CText, "before archive")

	// archive id 1
	_ = lysclient.MustDoToValue[string](t, srvApp.getRouter(), "DELETE", targetUrl+"/archive")

	// try to get id 1
	_, err := lysclient.DoToValueTester[string](srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, "row(s) not found", err.Error())

	// restore id 1
	_ = lysclient.MustDoToValue[string](t, srvApp.getRouter(), "POST", targetUrl+"/restore")

	// get id 1
	item = lysclient.MustDoToValue[corearchivetestm.Model](t, srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, 1, *item.CInt, "after archive")
	assert.Nil(t, item.CText, "after archive")
}

func TestArchiveByIdFailure(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	targetUrl := "/archive-test/100000"

	// archive invalid id
	_, err := lysclient.DoToValueTester[string](srvApp.getRouter(), "DELETE", targetUrl+"/archive")
	assert.EqualValues(t, "row(s) not found", err.Error(), "invalid id")

	// restore invalid id
	_, err = lysclient.DoToValueTester[string](srvApp.getRouter(), "POST", targetUrl+"/restore")
	assert.EqualValues(t, "row(s) not found", err.Error(), "invalid id")
}
