package lys

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/loveyourstack/lys/internal/stores/core/corearchivetest"
	"github.com/loveyourstack/lys/lysclient"
	"github.com/stretchr/testify/assert"
)

func TestArchiveByUuidSuccess(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// first, get id 1's uuid
	item := lysclient.MustDoToValue[corearchivetest.Model](t, srvApp.getRouter(), "GET", "/archive-test/1")

	targetUrl := "/archive-test-uuid/" + item.Iduu.String()

	// get id 1
	item = lysclient.MustDoToValue[corearchivetest.Model](t, srvApp.getRouter(), "GET", targetUrl)
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
	item = lysclient.MustDoToValue[corearchivetest.Model](t, srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, 1, *item.CInt, "after archive")
	assert.Nil(t, item.CText, "after archive")
}

func TestArchiveByUuidFailure(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	targetUrl := "/archive-test-uuid/" + uuid.New().String()

	// archive invalid id
	_, err := lysclient.DoToValueTester[string](srvApp.getRouter(), "DELETE", targetUrl+"/archive")
	assert.EqualValues(t, "row(s) not found", err.Error(), "invalid id")

	// restore invalid id
	_, err = lysclient.DoToValueTester[string](srvApp.getRouter(), "POST", targetUrl+"/restore")
	assert.EqualValues(t, "row(s) not found", err.Error(), "invalid id")
}
