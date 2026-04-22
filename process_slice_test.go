package lys

import (
	"context"
	"testing"

	"github.com/loveyourstack/lys/lysclient"
	"github.com/stretchr/testify/assert"
)

func TestProcessSliceSuccess(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	vals := []int64{1, 2}
	respVal := lysclient.MustPostToValue[[]int64, int64](ctx, t, srvApp.getRouter(), "POST", "/process-slice-test", vals)
	assert.EqualValues(t, int64(2), respVal)
}
