package lysrpc

import (
	"context"
	"testing"

	"github.com/loveyourstack/lys/internal/pb"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/stretchr/testify/assert"
)

func TestCreateEmptySuccess(t *testing.T) {

	// checks the result of posting a type test with as many fields missing as possible

	ctx := context.Background()
	conn := mustGetRpcConn(ctx, t)
	ttClient := pb.NewCoreTypeTestSvcClient(conn)

	input := coretypetestm.GetEmptyInput()
	item, err := ttClient.Create(ctx, CoreTTInputToPbFunc(input))
	if err != nil {
		t.Fatalf("ttClient.Create failed: %v", err)
	}

	// check nullable fields are nil
	assert.Nil(t, item.CBooln, "CBoolN")
	assert.Nil(t, item.CIntn, "CIntN")
	assert.Nil(t, item.CDoublen, "CDoubleN")
	assert.Nil(t, item.CNumericn, "CNumericN")
	assert.Nil(t, item.CEnumn, "CEnumN")
	assert.Nil(t, item.CTextn, "CTextN")

	// default lystype zeroable date types are empty string using RPC, not the zero values
	assert.EqualValues(t, "", item.CDate, "CDate")
	assert.EqualValues(t, "00:00", item.CTime, "CTime")
	assert.EqualValues(t, "", item.CDatetime, "CDatetime")
}

func TestPostFilledSuccess(t *testing.T) {

	// checks the result of posting a type test with all fields entered

	ctx := context.Background()
	conn := mustGetRpcConn(ctx, t)
	ttClient := pb.NewCoreTypeTestSvcClient(conn)

	input, err := coretypetestm.GetFilledInput()
	if err != nil {
		t.Fatalf("coretypetestm.GetFilledInput failed: %v", err)
	}

	item, err := ttClient.Create(ctx, CoreTTInputToPbFunc(input))
	if err != nil {
		t.Fatalf("ttClient.Create failed: %v", err)
	}

	coretypetestm.TestFilledInput(t, CoreTTModelFromPbFunc(item).Input)
}

func TestPostFailure(t *testing.T) {

	ctx := context.Background()
	conn := mustGetRpcConn(ctx, t)
	ttClient := pb.NewCoreTypeTestSvcClient(conn)

	// nil input (fails on mandatory enum val)
	_, err := ttClient.Create(ctx, nil)
	assert.ErrorContains(t, err, `invalid input value for enum core.weekday`, "nil")

	// empty struct (fails on mandatory enum val)
	_, err = ttClient.Create(ctx, CoreTTInputToPbFunc(coretypetestm.Input{}))
	assert.ErrorContains(t, err, `invalid input value for enum core.weekday`, "empty struct")
}
