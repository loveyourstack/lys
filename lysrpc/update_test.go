package lysrpc

import (
	"context"
	"testing"

	"github.com/loveyourstack/lys/internal/pb"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetest"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/loveyourstack/lys/internal/stores/core/coreuuidtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestUpdateSuccess(t *testing.T) {

	ctx := context.Background()
	conn := mustGetRpcConn(ctx, t)
	ttClient := pb.NewCoreTypeTestSvcClient(conn)

	// create a record with minimal values
	minInput := coretypetestm.GetEmptyInput()
	item, err := ttClient.Create(ctx, CoreTTInputToPbFunc(minInput))
	assert.NoError(t, err)

	// get filled type test input
	filledInput, err := coretypetestm.GetFilledInput()
	assert.NoError(t, err)

	// updated the filled input to the minimal record
	ttStore := coretypetest.Store{}
	item, err = ttClient.Update(ctx, &pb.UpdateCoreTypeTestReq{
		Id:    item.Id,
		Input: CoreTTInputToPbFunc(filledInput),
		FieldMask: &fieldmaskpb.FieldMask{
			Paths: ttStore.GetInputPlan().DbNames(),
		},
	})
	assert.NoError(t, err)

	// check changed record
	output, err := CoreTTModelFromPbFunc(item)
	require.NoError(t, err)
	coretypetestm.TestFilledInput(t, output.Input)
}

func TestUpdateUuid(t *testing.T) {

	ctx := context.Background()
	conn := mustGetRpcConn(ctx, t)
	utClient := pb.NewCoreUuidTestSvcClient(conn)

	input := coreuuidtest.Input{CInt: 1, CText: "a"}
	item, err := utClient.Create(ctx, CoreUTInputToPbFunc(input))
	assert.NoError(t, err)

	item, err = utClient.Update(ctx, &pb.UpdateCoreUuidTestReq{
		Id:    item.Id,
		Input: CoreUTInputToPbFunc(coreuuidtest.Input{CInt: 2, CText: "b"}),
		FieldMask: &fieldmaskpb.FieldMask{
			Paths: []string{"c_int", "c_text"},
		},
	})
	assert.NoError(t, err)

	assert.Equal(t, int64(2), item.CInt, "CInt")
	assert.Equal(t, "b", item.CText, "CText")
}

func TestUpdateFailure(t *testing.T) {

	ctx := context.Background()
	conn := mustGetRpcConn(ctx, t)
	ttClient := pb.NewCoreTypeTestSvcClient(conn)

	minInput := coretypetestm.GetEmptyInput()
	item, err := ttClient.Create(ctx, CoreTTInputToPbFunc(minInput))
	assert.NoError(t, err)

	t.Run("missing field mask", func(t *testing.T) {

		_, err := ttClient.Update(ctx, &pb.UpdateCoreTypeTestReq{
			Id:    item.Id,
			Input: CoreTTInputToPbFunc(coretypetestm.Input{CText: "a"}),
		})
		assert.EqualValues(t, "rpc error: code = InvalidArgument desc = field mask is required", err.Error(), "missing field mask")
	})

	t.Run("invalid field mask", func(t *testing.T) {

		_, err := ttClient.Update(ctx, &pb.UpdateCoreTypeTestReq{
			Id:    item.Id,
			Input: CoreTTInputToPbFunc(coretypetestm.Input{CText: "a"}),
			FieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"x"},
			},
		})
		assert.EqualValues(t, "rpc error: code = InvalidArgument desc = invalid field mask", err.Error(), "invalid field mask")
	})

	t.Run("id not found", func(t *testing.T) {

		_, err := ttClient.Update(ctx, &pb.UpdateCoreTypeTestReq{
			Id:    100000,
			Input: CoreTTInputToPbFunc(coretypetestm.Input{CText: "a"}),
			FieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"c_text"},
			},
		})
		assert.EqualValues(t, "rpc error: code = NotFound desc = id not found", err.Error(), "id not found")
	})
}
