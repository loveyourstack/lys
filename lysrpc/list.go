package lysrpc

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/loveyourstack/lys/lysapi"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lysset"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// iRpcListRequest is a request that can be used by RpcList.
// The type will typically be a protobuf message generated from a .proto file.
type iListRequest interface {
	GetFields() []string
	GetSorts() []string
	GetPage() int32
	GetPerPage() int32
	GetFilters() map[string]string
}

// List parses the request, selects items from the store, converts them using toPb and assembles the response using build.
// Note: if building req from JSON rather than a typed protobuf message, apply strict unmarshaling so that unknown fields are not ignored.
func List[modelT, pbModelT, pbRespT any](
	ctx context.Context,
	store lyspg.Selecter[modelT],
	req iListRequest,
	getOptions lysapi.GetOptions,
	toPb func(modelT) *pbModelT,
	build func(data []*pbModelT, meta lysapi.SelectMetadata) *pbRespT,
) (*pbRespT, error) {

	// parse request and select items from db
	items, unpagedCount, err := rpcSelect(ctx, store, req, getOptions)
	if err != nil {
		return nil, err // already a gRPC status error where appropriate
	}

	// convert db items to protobuf models
	data := make([]*pbModelT, 0, len(items))
	for _, item := range items {
		data = append(data, toPb(item))
	}

	// when using fields param, Postman by default still shows fields with zero values. Either change the Postman setting,
	// or use protojson.Marshal below to show the actual response to check that fields are omitted correctly.
	//b, _ := protojson.Marshal(data)
	//fmt.Println(string(b))

	// build the response using the provided build function and metadata
	return build(data, lysapi.SelectMetadata{
		Count:                 len(items),
		TotalCount:            unpagedCount.Value,
		TotalCountIsEstimated: unpagedCount.IsEstimated,
	}), nil
}

// rpcSelect parses the request, extracts fields, sorts, paging and filters, then selects items from the store.
func rpcSelect[T any](ctx context.Context, store lyspg.Selecter[T], req iListRequest, getOptions lysapi.GetOptions) (items []T, unpagedCount lyspg.TotalCount, err error) {

	dbNames := lysset.FromSlice(store.GetPlan().DbNames())
	jsonKeyDbNameMap := store.GetPlan().JsonKeyDbNameMap()

	// parse fields, if set
	fields, err := lysapi.ExtractFields(getOptions.FieldsParamName, strings.Join(req.GetFields(), ","), dbNames, jsonKeyDbNameMap)
	if err != nil {
		if userErr, ok := errors.AsType[lyserr.User](err); ok {
			return nil, lyspg.TotalCount{}, status.Error(codes.InvalidArgument, userErr.Error())
		}
		return nil, lyspg.TotalCount{}, fmt.Errorf("lys.ExtractFields failed: %w", err)
	}

	// parse sorts, if set
	sorts, err := lysapi.ExtractSorts(getOptions.SortParamName, strings.Join(req.GetSorts(), ","), jsonKeyDbNameMap)
	if err != nil {
		if userErr, ok := errors.AsType[lyserr.User](err); ok {
			return nil, lyspg.TotalCount{}, status.Error(codes.InvalidArgument, userErr.Error())
		}
		return nil, lyspg.TotalCount{}, fmt.Errorf("lys.ExtractSorts failed: %w", err)
	}

	// apply default paging if not set
	pageVal := req.GetPage()
	if pageVal == 0 {
		pageVal = 1
	}
	perPageVal := req.GetPerPage()
	if perPageVal == 0 {
		perPageVal = int32(getOptions.DefaultPerPage)
	}

	// parse paging
	page, perPage, err := lysapi.ExtractPaging(getOptions.PageParamName, fmt.Sprintf("%d", pageVal), getOptions.PerPageParamName,
		fmt.Sprintf("%d", perPageVal), getOptions.DefaultPerPage, getOptions.MaxPerPage)
	if err != nil {
		if userErr, ok := errors.AsType[lyserr.User](err); ok {
			return nil, lyspg.TotalCount{}, status.Error(codes.InvalidArgument, userErr.Error())
		}
		return nil, lyspg.TotalCount{}, fmt.Errorf("lys.ExtractPaging failed: %w", err)
	}

	// get offset from paging params (starts at 0, not 1)
	offset := perPage * (page - 1)

	// parse filters, if set

	// convert req map to url.Values for so lys library func can be used
	values := url.Values{}
	for k, v := range req.GetFilters() {
		values.Set(k, v)
	}

	conds, err := lysapi.ExtractFilters(values, jsonKeyDbNameMap, nil, nil, getOptions)
	if err != nil {
		if userErr, ok := errors.AsType[lyserr.User](err); ok {
			return nil, lyspg.TotalCount{}, status.Error(codes.InvalidArgument, userErr.Error())
		}
		return nil, lyspg.TotalCount{}, fmt.Errorf("lys.ExtractFilters failed: %w", err)
	}

	// select items from db
	items, unpagedCount, err = store.Select(ctx, lyspg.SelectParams{
		Fields:          fields,
		Conditions:      conds,
		Sorts:           sorts,
		Limit:           perPage,
		Offset:          offset,
		GetUnpagedCount: true,
	})
	if err != nil {
		return nil, lyspg.TotalCount{}, fmt.Errorf("store.Select failed: %w", err)
	}

	return items, unpagedCount, nil
}
