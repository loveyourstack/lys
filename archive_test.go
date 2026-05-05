package lys

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/loveyourstack/lys/internal/stores/core/corearchivetestm"
	"github.com/loveyourstack/lys/internal/stores/core/corearchivetestuuid"
	"github.com/loveyourstack/lys/lysclient"
	"github.com/stretchr/testify/assert"
)

func TestArchiveByIdSuccess(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	targetUrl := "/archive-test/1"

	// get id 1
	item := lysclient.MustDoToValue[corearchivetestm.Model](ctx, t, srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, 1, *item.CInt, "before archive")
	assert.Nil(t, item.CText, "before archive")

	// archive id 1
	_ = lysclient.MustDoToValue[string](ctx, t, srvApp.getRouter(), "DELETE", targetUrl+"/archive")

	// try to get id 1
	_, err := lysclient.DoToValueTester[string](ctx, srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, "row(s) not found", err.Error())

	// restore id 1
	_ = lysclient.MustDoToValue[string](ctx, t, srvApp.getRouter(), "POST", targetUrl+"/restore")

	// get id 1
	item = lysclient.MustDoToValue[corearchivetestm.Model](ctx, t, srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, 1, *item.CInt, "after archive")
	assert.Nil(t, item.CText, "after archive")
}

func TestArchiveByIdFailure(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	targetUrl := "/archive-test/100000"

	// archive invalid id
	_, err := lysclient.DoToValueTester[string](ctx, srvApp.getRouter(), "DELETE", targetUrl+"/archive")
	assert.EqualValues(t, "row(s) not found", err.Error(), "invalid id")

	// restore invalid id
	_, err = lysclient.DoToValueTester[string](ctx, srvApp.getRouter(), "POST", targetUrl+"/restore")
	assert.EqualValues(t, "row(s) not found", err.Error(), "invalid id")
}

func TestArchiveByUuidSuccess(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	targetUrl := "/archive-test-uuid/b872908f-68ca-4c87-96dd-98a9b97470f0"

	// get first record
	item := lysclient.MustDoToValue[corearchivetestuuid.Model](ctx, t, srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, 1, *item.CInt, "before archive")
	assert.Nil(t, item.CText, "before archive")

	// archive first record
	_ = lysclient.MustDoToValue[string](ctx, t, srvApp.getRouter(), "DELETE", targetUrl+"/archive")

	// try to get first record
	_, err := lysclient.DoToValueTester[string](ctx, srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, "row(s) not found", err.Error())

	// restore first record
	_ = lysclient.MustDoToValue[string](ctx, t, srvApp.getRouter(), "POST", targetUrl+"/restore")

	// get first record
	item = lysclient.MustDoToValue[corearchivetestuuid.Model](ctx, t, srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, 1, *item.CInt, "after archive")
	assert.Nil(t, item.CText, "after archive")
}

func TestArchiveByUuidFailure(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	targetUrl := "/archive-test-uuid/" + uuid.New().String()

	// archive invalid id
	_, err := lysclient.DoToValueTester[string](ctx, srvApp.getRouter(), "DELETE", targetUrl+"/archive")
	assert.EqualValues(t, "row(s) not found", err.Error(), "invalid id")

	// restore invalid id
	_, err = lysclient.DoToValueTester[string](ctx, srvApp.getRouter(), "POST", targetUrl+"/restore")
	assert.EqualValues(t, "row(s) not found", err.Error(), "invalid id")
}
