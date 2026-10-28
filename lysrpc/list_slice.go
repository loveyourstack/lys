package lysrpc

import "github.com/loveyourstack/lys/lysapi"

// ToPbSlice converts items to protobuf models using toPb. The result is never nil.
func ToPbSlice[modelT, pbModelT any](items []modelT, toPb func(modelT) *pbModelT) []*pbModelT {
	data := make([]*pbModelT, len(items))
	for i, item := range items {
		data[i] = toPb(item)
	}
	return data
}

// ListSlice builds a list response from an in-memory slice (e.g. sessions or login attempts) that is not selected from a store.
// All items are returned in one page, so Count and TotalCount are both len(items).
func ListSlice[modelT, pbModelT, pbRespT any](
	items []modelT,
	toPb func(modelT) *pbModelT,
	build func(data []*pbModelT, meta lysapi.SelectMetadata) *pbRespT,
) *pbRespT {
	return build(ToPbSlice(items, toPb), lysapi.SelectMetadata{
		Count:      len(items),
		TotalCount: int64(len(items)),
	})
}
