package lystype

import (
	"testing"
	"time"

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

	expected0 := map[string]any{"id": 1, "name": "one"}
	expected1 := map[string]any{"id": 2, "name": "two"}

	assert.Equal(t, expected0, recsMap[0], "record 0")
	assert.Equal(t, expected1, recsMap[1], "record 1")
}

func TestRecsToMapPointerRecordsSuccess(t *testing.T) {

	type recS struct {
		ID   int     `json:"id"`
		Note *string `json:"note"`
	}

	note := "hello"
	recs := []*recS{
		{ID: 1, Note: &note},
		{ID: 2, Note: nil},
	}

	recsMap, err := RecsToMap(recs)
	assert.NoError(t, err)
	assert.Len(t, recsMap, 2)
	assert.Equal(t, map[string]any{"id": 1, "note": "hello"}, recsMap[0])
	assert.Equal(t, map[string]any{"id": 2, "note": nil}, recsMap[1])
}

func TestRecsToMapOmitOptions(t *testing.T) {

	type recS struct {
		Name string `json:"name,omitempty"`
		DOB  Date   `json:"dob,omitzero"`
		City string `json:"city"`
	}

	d := Date(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	recs := []recS{
		{Name: "", DOB: Date{}, City: "x"},
		{Name: "ann", DOB: d, City: "y"},
	}

	recsMap, err := RecsToMap(recs)
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{"city": "x"}, recsMap[0])
	assert.Equal(t, map[string]any{"name": "ann", "dob": d, "city": "y"}, recsMap[1])
}

func TestRecsToMapAnonymousEmbeddedFlattening(t *testing.T) {

	type innerS struct {
		Code string `json:"code"`
	}
	type recS struct {
		innerS
		Name string `json:"name"`
	}

	recs := []recS{{innerS: innerS{Code: "A"}, Name: "alpha"}}

	recsMap, err := RecsToMap(recs)
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{"code": "A", "name": "alpha"}, recsMap[0])
}

func TestRecsToMapKeepsNativeCustomType(t *testing.T) {

	type recS struct {
		Start Date `json:"start"`
	}

	d := Date(time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC))
	recs := []recS{{Start: d}}

	recsMap, err := RecsToMap(recs)
	assert.NoError(t, err)

	val, ok := recsMap[0]["start"].(Date)
	assert.True(t, ok)
	assert.Equal(t, d, val)
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

func TestRecsToMapFirstElementTypeFailure(t *testing.T) {
	recs := []int{1, 2, 3}

	_, err := RecsToMap(recs)
	assert.EqualError(t, err, "T must be a struct or pointer to struct")
}

func TestRecsToMapNilPointerFailure(t *testing.T) {

	type recS struct {
		ID int `json:"id"`
	}

	recs := []*recS{{ID: 1}, nil}

	_, err := RecsToMap(recs)
	assert.EqualError(t, err, "recs[1] is nil")
}
