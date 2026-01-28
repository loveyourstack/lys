package lystype

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecsToMapSuccess(t *testing.T) {

	type recS struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Excluded  string `json:"-"`
		NoJsonTag string
	}

	recs := []recS{
		{ID: 1, Name: "one", Excluded: "excluded1", NoJsonTag: "nojsontag1"},
		{ID: 2, Name: "two", Excluded: "excluded2", NoJsonTag: "nojsontag2"},
	}

	recsMap, err := RecsToMap(recs)
	assert.NoError(t, err, "RecsToMap should not error")
	assert.Len(t, recsMap, 2, "recsMap length")

	expected0 := map[string]any{"id": float64(1), "name": "one", "NoJsonTag": "nojsontag1"}
	expected1 := map[string]any{"id": float64(2), "name": "two", "NoJsonTag": "nojsontag2"}

	assert.Equal(t, expected0, recsMap[0], "record 0")
	assert.Equal(t, expected1, recsMap[1], "record 1")
}

func TestRecsToMapEmptyFailure(t *testing.T) {

	type recS struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	recs := []recS{}

	_, err := RecsToMap(recs)
	assert.Error(t, err, "empty recs")
}
