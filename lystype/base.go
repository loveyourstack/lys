package lystype

import (
	"encoding/json"
	"fmt"
)

// RecsToMap converts a slice of T to a map[string]any using JSON marshal/unmarshal
// only those T fields with a json tag will be in the result
func RecsToMap[T any](recs []T) (recsMap []map[string]any, err error) {

	if len(recs) == 0 {
		return nil, fmt.Errorf("recs is empty")
	}

	jsonRecs, err := json.Marshal(recs)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal failed: %w", err)
	}

	err = json.Unmarshal(jsonRecs, &recsMap)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	return recsMap, nil
}

func ToPtr[T any](a T) *T {
	return &a
}
