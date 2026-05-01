package lysclient

import "net/http"

// structs and funcs for testing lys endpoints which return StdResponse

// GetMetadata contains the metadata for a GET request which returns a slice of items (db records)
type GetMetadata struct {
	Count                 int   `json:"count"`
	TotalCount            int64 `json:"total_count"`
	TotalCountIsEstimated bool  `json:"total_count_is_estimated"`
}

// ItemSResp is expected when StdResponse returns a slice of map[string]any
type ItemSResp struct {
	Status         string           `json:"status"`
	Data           []map[string]any `json:"data"`
	GetMetadata    *GetMetadata     `json:"metadata,omitempty"` // only used for GET many
	ErrDescription string           `json:"err_description"`
}

// SliceResp is expected when StdResponse returns a slice of T
type SliceResp[T any] struct {
	Status         string `json:"status"`
	Data           []T    `json:"data"`
	ErrDescription string `json:"err_description"`
}

// ValueResp is expected when StdResponse returns a T
type ValueResp[T any] struct {
	Status         string `json:"status"`
	Data           T      `json:"data"`
	ErrDescription string `json:"err_description"`
}

const successStatus string = "succeeded"

var allowedPostMethods = []string{http.MethodPost, http.MethodPut, http.MethodPatch}
