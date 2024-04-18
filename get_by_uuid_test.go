package lys

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetest"
	"github.com/loveyourstack/lys/lysclient"
	"github.com/stretchr/testify/assert"
)

func TestGetByUuidSuccess(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// first, get id 1
	targetUrl := "/type-test/1"
	item := lysclient.MustDoToValue[coretypetest.Model](t, srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, true, item.CBool)

	// use id 1's uuid
	targetUrl = "/type-test-uuid/" + item.Iduu.String()
	itemUuid := lysclient.MustDoToValue[coretypetest.Model](t, srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, true, itemUuid.CBool)
}

func TestGetByUuidFailure(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// uuid invalid
	targetUrl := "/type-test-uuid/a"
	_, err := lysclient.DoToValueTester[coretypetest.Model](srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, "id not a uuid", err.Error())

	// uuid doesn't exist
	targetUrl = "/type-test-uuid/" + uuid.New().String()
	_, err = lysclient.DoToValueTester[coretypetest.Model](srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, "invalid id", err.Error())
}
