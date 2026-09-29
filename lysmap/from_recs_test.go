package lysmap

import (
	"testing"
	"time"

	"github.com/loveyourstack/lys/lystype"
	"github.com/stretchr/testify/assert"
)

func TestFromRecsSuccess(t *testing.T) {

	t.Run("with recs", func(t *testing.T) {
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

		recsMap, err := FromRecs(recs)
		assert.NoError(t, err, "FromRecs should not error")
		assert.Len(t, recsMap, 2, "recsMap length")

		expected0 := map[string]any{"id": 1, "name": "one"}
		expected1 := map[string]any{"id": 2, "name": "two"}

		assert.Equal(t, expected0, recsMap[0], "record 0")
		assert.Equal(t, expected1, recsMap[1], "record 1")
	})

	t.Run("empty recs", func(t *testing.T) {
		type recS struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}

		recs := []recS{}

		_, err := FromRecs(recs)
		assert.NoError(t, err)
	})
}

func TestFromRecsPointerRecordsSuccess(t *testing.T) {

	type recS struct {
		ID   int     `json:"id"`
		Note *string `json:"note"`
	}

	note := "hello"
	recs := []*recS{
		{ID: 1, Note: &note},
		{ID: 2, Note: nil},
	}

	recsMap, err := FromRecs(recs)
	assert.NoError(t, err)
	assert.Len(t, recsMap, 2)
	assert.Equal(t, map[string]any{"id": 1, "note": "hello"}, recsMap[0])
	assert.Equal(t, map[string]any{"id": 2, "note": nil}, recsMap[1])
}

func TestFromRecsEmptyJSONTagNameUsesFieldName(t *testing.T) {
	type recS struct {
		Name string `json:",omitzero"`
	}

	recMap, err := FromRecs([]recS{{Name: "ann"}})

	assert.NoError(t, err)
	assert.Equal(t, map[string]any{"Name": "ann"}, recMap[0])
}

func TestFromRecsOmitEmptyNilValues(t *testing.T) {
	type recS struct {
		Pointer  *string `json:"pointer,omitempty"`
		Any      any     `json:"any,omitempty"`
		TypedNil any     `json:"typed_nil,omitempty"`
	}

	var pointer *string
	var typedNil *string

	recMap, err := FromRecs([]recS{{
		Pointer:  pointer,
		Any:      nil,
		TypedNil: typedNil,
	}})

	assert.NoError(t, err)
	assert.Empty(t, recMap[0])
}

func TestFromRecsOmitOptions(t *testing.T) {

	t.Run("regular types", func(t *testing.T) {
		type recS struct {
			Name string       `json:"name,omitzero"`
			DOB  lystype.Date `json:"dob,omitzero"`
			City string       `json:"city"`
		}

		d := lystype.Date(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
		recs := []recS{
			{Name: "", DOB: lystype.Date{}, City: "x"},
			{Name: "ann", DOB: d, City: "y"},
		}

		recsMap, err := FromRecs(recs)
		assert.NoError(t, err)
		assert.Equal(t, map[string]any{"city": "x"}, recsMap[0])
		assert.Equal(t, map[string]any{"name": "ann", "dob": d, "city": "y"}, recsMap[1])
	})

	t.Run("nil and empty types", func(t *testing.T) {
		type recS struct {
			EmptySlice     []string          `json:"empty_slice,omitzero"`
			NilSlice       []string          `json:"nil_slice,omitzero"`
			EmptySliceOmit []string          `json:"empty_slice_omit,omitempty"`
			EmptyMap       map[string]string `json:"empty_map,omitzero"`
			NilMap         map[string]string `json:"nil_map,omitzero"`
			ZeroArray      [1]int            `json:"zero_array,omitzero"`
		}

		recsMap, err := FromRecs([]recS{{
			EmptySlice:     []string{},
			NilSlice:       nil,
			EmptySliceOmit: []string{},
			EmptyMap:       map[string]string{},
			NilMap:         nil,
			ZeroArray:      [1]int{0},
		}})

		assert.NoError(t, err)
		assert.Equal(t, map[string]any{
			"empty_slice": []string{},
			"empty_map":   map[string]string{},
		}, recsMap[0])
	})
}

type semanticValue int

func (v semanticValue) IsZero() bool {
	return v == 42
}

func TestFromRecsOmitZeroUsesIsZero(t *testing.T) {
	type recS struct {
		Value semanticValue `json:"value,omitzero"`
	}

	recMap, err := FromRecs([]recS{{Value: 42}})

	assert.NoError(t, err)
	assert.Empty(t, recMap[0])
}

func TestFromRecsEmbeddedFlattening(t *testing.T) {

	type innerS struct {
		Code string `json:"code"`
	}
	type recS struct {
		innerS
		Name string `json:"name"`
	}

	recs := []recS{{Code: "A", Name: "alpha"}}

	recsMap, err := FromRecs(recs)
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{"code": "A", "name": "alpha"}, recsMap[0])
}

func TestFromRecsKeepsNativeCustomType(t *testing.T) {

	type recS struct {
		Start lystype.Date `json:"start"`
	}

	d := lystype.Date(time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC))
	recs := []recS{{Start: d}}

	recsMap, err := FromRecs(recs)
	assert.NoError(t, err)

	val, ok := recsMap[0]["start"].(lystype.Date)
	assert.True(t, ok)
	assert.Equal(t, d, val)
}

func TestFromRecsFirstElementTypeFailure(t *testing.T) {
	recs := []int{1, 2, 3}

	_, err := FromRecs(recs)
	assert.EqualError(t, err, "T must be a struct or pointer to struct")
}

func TestFromRecsNilPointerFailure(t *testing.T) {

	type recS struct {
		ID int `json:"id"`
	}

	recs := []*recS{{ID: 1}, nil}

	_, err := FromRecs(recs)
	assert.EqualError(t, err, "recs[1] is nil")
}
