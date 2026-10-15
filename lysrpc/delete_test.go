package lysrpc

import (
	"context"
	"testing"

	"github.com/loveyourstack/lys/internal/pb"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/loveyourstack/lys/internal/stores/core/coreuuidtest"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDeleteSuccess(t *testing.T) {

	ctx := context.Background()
	conn := mustGetRpcConn(ctx, t)
	ttClient := pb.NewCoreTypeTestSvcClient(conn)

	// create record
	minInput := coretypetestm.GetEmptyInput()
	item, err := ttClient.Create(ctx, CoreTTInputToPbFunc(minInput))
	assert.NoError(t, err)

	// delete it
	_, err = ttClient.Delete(ctx, &pb.DeleteByIdReq{Id: item.Id})
	assert.NoError(t, err)

	// try to select it
	_, err = ttClient.Get(ctx, &pb.GetByIdReq{Id: item.Id})
	assert.EqualValues(t, "rpc error: code = NotFound desc = item not found", err.Error())
}

func TestDeleteUuid(t *testing.T) {

	ctx := context.Background()
	conn := mustGetRpcConn(ctx, t)
	utClient := pb.NewCoreUuidTestSvcClient(conn)

	// create record
	input := coreuuidtest.Input{
		CInt:  4,
		CText: "d",
	}
	item, err := utClient.Create(ctx, CoreUTInputToPbFunc(input))
	assert.NoError(t, err)

	// delete it
	_, err = utClient.Delete(ctx, &pb.DeleteByUuidReq{Id: item.Id})
	assert.NoError(t, err)

	// try to select it
	_, err = utClient.Get(ctx, &pb.GetByUuidReq{Id: item.Id})
	assert.EqualValues(t, "rpc error: code = NotFound desc = item not found", err.Error())
}

func TestDeleteFailure(t *testing.T) {

	ctx := context.Background()
	conn := mustGetRpcConn(ctx, t)
	ttClient := pb.NewCoreTypeTestSvcClient(conn)

	// id doesn't exist
	_, err := ttClient.Delete(ctx, &pb.DeleteByIdReq{Id: 100000})
	assert.EqualValues(t, "rpc error: code = NotFound desc = item not found", err.Error())
	assert.Equal(t, codes.NotFound, status.Code(err))
}
