package lysrpc

import (
	"context"
	"testing"
	"uuid"

	"github.com/loveyourstack/lys/internal/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetByIdSuccess(t *testing.T) {

	ctx := context.Background()
	conn := mustGetRpcConn(ctx, t)
	ttClient := pb.NewCoreTypeTestSvcClient(conn)

	item, err := ttClient.Get(ctx, &pb.GetByIdReq{Id: 1})
	assert.NoError(t, err)

	assert.EqualValues(t, true, item.CBool)
	assert.EqualValues(t, "a b", item.CText)
}

func TestGetByIdFailure(t *testing.T) {

	ctx := context.Background()
	conn := mustGetRpcConn(ctx, t)
	ttClient := pb.NewCoreTypeTestSvcClient(conn)

	// id doesn't exist
	_, err := ttClient.Get(ctx, &pb.GetByIdReq{Id: 100000})
	assert.EqualValues(t, "rpc error: code = NotFound desc = item not found", err.Error())
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestGetByUuidSuccess(t *testing.T) {

	ctx := context.Background()
	conn := mustGetRpcConn(ctx, t)
	utClient := pb.NewCoreUuidTestSvcClient(conn)

	// check first record
	item, err := utClient.Get(ctx, &pb.GetByUuidReq{Id: "b872908f-68ca-4c87-96dd-98a9b97470f0"})
	assert.NoError(t, err)

	assert.EqualValues(t, 1, item.CInt)
	assert.EqualValues(t, "a", item.CText)
}

func TestGetByUuidFailure(t *testing.T) {

	ctx := context.Background()
	conn := mustGetRpcConn(ctx, t)
	utClient := pb.NewCoreUuidTestSvcClient(conn)

	// uuid invalid
	_, err := utClient.Get(ctx, &pb.GetByUuidReq{Id: "a"})
	assert.EqualValues(t, "rpc error: code = InvalidArgument desc = invalid uuid", err.Error())
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	// uuid doesn't exist
	_, err = utClient.Get(ctx, &pb.GetByUuidReq{Id: uuid.New().String()})
	assert.EqualValues(t, "rpc error: code = NotFound desc = item not found", err.Error())
	assert.Equal(t, codes.NotFound, status.Code(err))
}
