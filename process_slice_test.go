package lys

import (
	"context"
	"testing"

	"github.com/loveyourstack/lys/lysclient"
	"github.com/stretchr/testify/assert"
)

func TestProcessSliceSuccess(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	vals := []int64{1, 2}
	respVal := lysclient.MustPostToValue[[]int64, int64](t, srvApp.getRouter(), "POST", "/process-slice-test", vals)
	assert.EqualValues(t, int64(2), respVal)
}
