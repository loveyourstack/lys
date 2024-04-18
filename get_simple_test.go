package lys

import (
	"context"
	"testing"

	"github.com/loveyourstack/lys/lysclient"
	"github.com/stretchr/testify/assert"
)

func TestGetSimple(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// call route handled by GetSimple
	targetUrl := "/volume-test/any-10"
	vals := lysclient.MustGetArray[int](t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, 10, len(vals))
}
