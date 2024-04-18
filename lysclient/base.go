package lysclient

// structs and funcs for testing lys endpoints which return StdResponse

// ArrayResp is expected when StdResponse returns an array of T
type ArrayResp[T any] struct {
	Status         string `json:"status"`
	Data           []T    `json:"data"`
	ErrDescription string `json:"err_description"`
}

// ItemAResp is expected when StdResponse returns an array of items (db records)
type ItemAResp struct {
	Status         string           `json:"status"`
	Data           []map[string]any `json:"data"`
	ErrDescription string           `json:"err_description"`
}

// ValueResp is expected when StdResponse returns a T
type ValueResp[T any] struct {
	Status         string `json:"status"`
	Data           T      `json:"data"`
	ErrDescription string `json:"err_description"`
}

const successStatus string = "succeeded"
