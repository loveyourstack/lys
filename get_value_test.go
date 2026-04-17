package lys

import (
	"context"
	"testing"

	"github.com/loveyourstack/lys/lysclient"
	"github.com/stretchr/testify/assert"
)

func TestGetValue(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// call route handled by GetValue
	targetUrl := "/volume-test/int-1"
	val := lysclient.MustGetValue[int](t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, 1, val)
}
