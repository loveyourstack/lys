package lysrpc

import (
	"context"
	"testing"

	"github.com/loveyourstack/lys/internal/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestListSuccess(t *testing.T) {

	ctx := context.Background()
	conn := mustGetRpcConn(ctx, t)
	ptClient := pb.NewCoreParamTestSvcClient(conn)

	items, err := ptClient.List(ctx, nil)
	assert.NoError(t, err)

	// basic test only: sorting, filtering etc is tested below
	assert.EqualValues(t, false, items.Metadata.TotalCountIsEstimated)
	assert.EqualValues(t, 2, items.Metadata.TotalCount)
	assert.EqualValues(t, 2, items.Metadata.Count)
	assert.EqualValues(t, 2, len(items.Data))
}

func TestListFailure(t *testing.T) {

	ctx := context.Background()
	conn := mustGetRpcConn(ctx, t)
	ptClient := pb.NewCoreParamTestSvcClient(conn)

	// invalid filter field
	_, err := ptClient.List(ctx, &pb.ListReq{
		Filters: map[string]string{"a": "b"},
	})
	assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = invalid filter field: a")
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestListSuccessFields(t *testing.T) {

	ctx := context.Background()
	conn := mustGetRpcConn(ctx, t)
	ptClient := pb.NewCoreParamTestSvcClient(conn)

	// select id only
	items, err := ptClient.List(ctx, &pb.ListReq{
		Fields: []string{"id"},
	})
	assert.NoError(t, err)
	item := items.Data[0]

	// present
	if item.Id == 0 {
		t.Errorf("id: expected in response, but missing")
	}

	// not present
	if item.CInt != 0 {
		t.Errorf("c_int: not expected in response, but present")
	}
	if item.CDate != "" {
		t.Errorf("c_date: not expected in response, but present")
	}
	if item.CTime != "" {
		t.Errorf("c_time: not expected in response, but present")
	}
	if item.CDatetime != "" {
		t.Errorf("c_datetime: not expected in response, but present")
	}
}

func TestListSuccessFilters(t *testing.T) {

	ctx := context.Background()
	conn := mustGetRpcConn(ctx, t)
	ptClient := pb.NewCoreParamTestSvcClient(conn)

	// each filter string should return 1 item
	filterStrA := []map[string]string{

		// equals
		{"c_bool": "true"},
		{"c_int": "1"},
		{"c_double": "1.1"},
		{"c_date": "2001-01-01"},
		{"c_time": "12:01"},
		{"c_datetime": "2001-01-01 12:01:00+01"},
		{"c_enum": "Monday"},
		{"c_text": "a"},
		{"c_textn": "{empty}"},
		{"c_booln": "{null}"},

		// not equals
		{"c_int": "!1"},
		{"c_double": "!1.1"},
		{"c_date": "!2001-01-01"},
		{"c_time": "!12:01"},
		{"c_datetime": "!2001-01-01 12:01:00+01"},
		{"c_enum": "!Monday"},
		{"c_text": "!a"},
		{"c_textn": "{!empty}"},
		{"c_booln": "{!null}"},

		// greater than
		{"c_int": ">1"},
		{"c_double": ">1.1"},
		{"c_date": ">2001-01-01"},
		{"c_time": ">12:01"},
		{"c_datetime": ">2001-01-01 12:01:00+01"},

		// less than
		{"c_int": "<2"},
		{"c_double": "<2.1"},
		{"c_date": "<2002-01-01"},
		{"c_time": "<12:02"},
		{"c_datetime": "<2002-01-01 12:01:00+01"},

		// starts with
		{"c_date": "2001~"},
		{"c_time": "12:01~"},
		{"c_datetime": "2001~"},
		{"c_textn": "a~"},

		// ends with
		{"c_date": "~01-01-01"},
		{"c_time": "~01:00"},
		{"c_datetime": "~01-01-01 12:01:00+01"},
		{"c_textn": "~c"},

		// contains
		{"c_date": "~001~"},
		{"c_time": "~2:01~"},
		{"c_datetime": "~001~"},
		{"c_textn": "~b~"},

		// not contains
		{"c_textn": "!~b~"},

		// in
		{"c_textn": "abc|d"},

		// not in
		{"c_textn": "!abc|d"},

		// contains any
		{"c_textn": "~[b|d]~"},
	}

	for _, filterStr := range filterStrA {
		items, err := ptClient.List(ctx, &pb.ListReq{Filters: filterStr})
		assert.NoError(t, err)
		assert.EqualValues(t, 1, len(items.Data), filterStr)
	}
}

func TestListSuccessPaging(t *testing.T) {

	ctx := context.Background()
	conn := mustGetRpcConn(ctx, t)
	ptClient := pb.NewCoreParamTestSvcClient(conn)

	// page 1, limit 1
	items, err := ptClient.List(ctx, &pb.ListReq{Page: 1, PerPage: 1, Sorts: []string{"c_text"}})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(items.Data), "page 1 limit 1 len")
	assert.EqualValues(t, "a", items.Data[0].CText, "page 1 limit 1 value")

	// page 2, limit 1

	items, err = ptClient.List(ctx, &pb.ListReq{Page: 2, PerPage: 1, Sorts: []string{"c_text"}})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(items.Data), "page 2 limit 1 len")
	assert.EqualValues(t, "b", items.Data[0].CText, "page 2 limit 1 value")
}

func TestListSuccessSorts(t *testing.T) {

	ctx := context.Background()
	conn := mustGetRpcConn(ctx, t)
	ptClient := pb.NewCoreParamTestSvcClient(conn)

	// ascending: each sort string should return the first record
	sortStrA := [][]string{
		{"c_text"},
		{"c_intn", "c_text"},
	}
	for _, sortStr := range sortStrA {
		items, err := ptClient.List(ctx, &pb.ListReq{Sorts: sortStr})
		assert.NoError(t, err)
		assert.EqualValues(t, "a", items.Data[0].CText, sortStr)
	}

	// descending: each sort string should return the second record
	sortStrA = [][]string{
		{"-c_text"},
		{"c_intn", "-c_text"},
	}
	for _, sortStr := range sortStrA {
		items, err := ptClient.List(ctx, &pb.ListReq{Sorts: sortStr})
		assert.NoError(t, err)
		assert.EqualValues(t, "b", items.Data[0].CText, sortStr)
	}
}
