package lysmap

import (
	"testing"
	"time"

	"github.com/loveyourstack/lys/lystype"
	"github.com/stretchr/testify/assert"
)

func TestFromItemsSuccess(t *testing.T) {

	t.Run("with items", func(t *testing.T) {
		type itemS struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			Excluded  string `json:"-"`
			NoJsonTag string
		}

		items := []itemS{
			{ID: 1, Name: "one", Excluded: "excluded1", NoJsonTag: "nojsontag1"},
			{ID: 2, Name: "two", Excluded: "excluded2", NoJsonTag: "nojsontag2"},
		}

		itemMap, err := FromItems(items)
		assert.NoError(t, err, "FromItems should not error")
		assert.Len(t, itemMap, 2, "itemMap length")

		expected0 := map[string]any{"id": 1, "name": "one"}
		expected1 := map[string]any{"id": 2, "name": "two"}

		assert.Equal(t, expected0, itemMap[0], "item 0")
		assert.Equal(t, expected1, itemMap[1], "item 1")
	})

	t.Run("empty items", func(t *testing.T) {
		type itemS struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}

		items := []itemS{}

		_, err := FromItems(items)
		assert.NoError(t, err)
	})
}

func TestFromItemsPointerSuccess(t *testing.T) {

	type itemS struct {
		ID   int     `json:"id"`
		Note *string `json:"note"`
	}

	note := "hello"
	items := []*itemS{
		{ID: 1, Note: &note},
		{ID: 2, Note: nil},
	}

	itemsMap, err := FromItems(items)
	assert.NoError(t, err)
	assert.Len(t, itemsMap, 2)
	assert.Equal(t, map[string]any{"id": 1, "note": "hello"}, itemsMap[0])
	assert.Equal(t, map[string]any{"id": 2, "note": nil}, itemsMap[1])
}

func TestFromItemsEmptyJSONTagNameUsesFieldName(t *testing.T) {
	type itemS struct {
		Name string `json:",omitzero"`
	}

	itemMap, err := FromItems([]itemS{{Name: "ann"}})

	assert.NoError(t, err)
	assert.Equal(t, map[string]any{"Name": "ann"}, itemMap[0])
}

func TestFromItemsOmitEmptyNilValues(t *testing.T) {
	type itemS struct {
		Pointer  *string `json:"pointer,omitempty"`
		Any      any     `json:"any,omitempty"`
		TypedNil any     `json:"typed_nil,omitempty"`
	}

	var pointer *string
	var typedNil *string

	itemMap, err := FromItems([]itemS{{
		Pointer:  pointer,
		Any:      nil,
		TypedNil: typedNil,
	}})

	assert.NoError(t, err)
	assert.Empty(t, itemMap[0])
}

func TestFromItemsOmitOptions(t *testing.T) {

	t.Run("regular types", func(t *testing.T) {
		type itemS struct {
			Name string       `json:"name,omitzero"`
			DOB  lystype.Date `json:"dob,omitzero"`
			City string       `json:"city"`
		}

		d := lystype.Date(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
		items := []itemS{
			{Name: "", DOB: lystype.Date{}, City: "x"},
			{Name: "ann", DOB: d, City: "y"},
		}

		itemMap, err := FromItems(items)
		assert.NoError(t, err)
		assert.Equal(t, map[string]any{"city": "x"}, itemMap[0])
		assert.Equal(t, map[string]any{"name": "ann", "dob": d, "city": "y"}, itemMap[1])
	})

	t.Run("nil and empty types", func(t *testing.T) {
		type itemS struct {
			EmptySlice     []string          `json:"empty_slice,omitzero"`
			NilSlice       []string          `json:"nil_slice,omitzero"`
			EmptySliceOmit []string          `json:"empty_slice_omit,omitempty"`
			EmptyMap       map[string]string `json:"empty_map,omitzero"`
			NilMap         map[string]string `json:"nil_map,omitzero"`
			ZeroArray      [1]int            `json:"zero_array,omitzero"`
		}

		itemMap, err := FromItems([]itemS{{
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
		}, itemMap[0])
	})
}

type semanticValue int

func (v semanticValue) IsZero() bool {
	return v == 42
}

func TestFromItemsOmitZeroUsesIsZero(t *testing.T) {
	type itemS struct {
		Value semanticValue `json:"value,omitzero"`
	}

	itemMap, err := FromItems([]itemS{{Value: 42}})

	assert.NoError(t, err)
	assert.Empty(t, itemMap[0])
}

func TestFromItemsEmbeddedFlattening(t *testing.T) {

	type innerS struct {
		Code string `json:"code"`
	}
	type itemS struct {
		innerS
		Name string `json:"name"`
	}

	items := []itemS{{Code: "A", Name: "alpha"}}

	itemMap, err := FromItems(items)
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{"code": "A", "name": "alpha"}, itemMap[0])
}

func TestFromItemsKeepsNativeCustomType(t *testing.T) {

	type itemS struct {
		Start lystype.Date `json:"start"`
	}

	d := lystype.Date(time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC))
	items := []itemS{{Start: d}}

	itemMap, err := FromItems(items)
	assert.NoError(t, err)

	val, ok := itemMap[0]["start"].(lystype.Date)
	assert.True(t, ok)
	assert.Equal(t, d, val)
}

func TestFromItemsFirstElementTypeFailure(t *testing.T) {
	items := []int{1, 2, 3}

	_, err := FromItems(items)
	assert.EqualError(t, err, "T must be a struct or pointer to struct")
}

func TestFromItemsNilPointerFailure(t *testing.T) {

	type itemS struct {
		ID int `json:"id"`
	}

	items := []*itemS{{ID: 1}, nil}

	_, err := FromItems(items)
	assert.EqualError(t, err, "items[1] is nil")
}
