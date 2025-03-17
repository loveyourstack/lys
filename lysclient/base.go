package lysclient

// structs and funcs for testing lys endpoints which return StdResponse

// ArrayResp is expected when StdResponse returns an array of T
type ArrayResp[T any] struct {
	Status         string `json:"status"`
	Data           []T    `json:"data"`
	ErrDescription string `json:"err_description"`
}

type GetMetadata struct {
	Count                 int   `json:"count"`
	TotalCount            int64 `json:"total_count"`
	TotalCountIsEstimated bool  `json:"total_count_is_estimated"`
}

// ItemAResp is expected when StdResponse returns an array of items (db records)
type ItemAResp struct {
	Status         string           `json:"status"`
	Data           []map[string]any `json:"data"`
	GetMetadata    *GetMetadata     `json:"metadata,omitempty"` // only used for GET many
	ErrDescription string           `json:"err_description"`
}

// ValueResp is expected when StdResponse returns a T
type ValueResp[T any] struct {
	Status         string `json:"status"`
	Data           T      `json:"data"`
	ErrDescription string `json:"err_description"`
}

const successStatus string = "succeeded"
