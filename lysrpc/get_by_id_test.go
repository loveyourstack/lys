package lysrpc

import (
	"context"
	"testing"

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
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

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
