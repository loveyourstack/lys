package lysapi

type SelectMetadata struct {
	Count                 int   `json:"count"`
	TotalCount            int64 `json:"total_count"`
	TotalCountIsEstimated bool  `json:"total_count_is_estimated"`
}
