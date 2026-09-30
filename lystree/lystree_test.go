package lystree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testItem struct {
	ID       int
	ParentID int
	Name     string
}

func testItemID(item testItem) int {
	return item.ID
}

func testItemParentID(item testItem) int {
	return item.ParentID
}

func TestFromItemsBuildsTree(t *testing.T) {
	items := []testItem{
		{ID: 2, ParentID: 1, Name: "child"},
		{ID: 3, ParentID: 2, Name: "grandchild"},
		{ID: 1, ParentID: 1, Name: "root"},
		{ID: 4, ParentID: 4, Name: "second root"},
	}

	nodes, err := FromItems(items, func(item testItem) int {
		return item.ID
	}, func(item testItem) int {
		return item.ParentID
	})

	assert.NoError(t, err)
	if !assert.Len(t, nodes, 2) {
		return
	}

	assert.Equal(t, 1, nodes[0].Item.ID)
	assert.Equal(t, "root", nodes[0].Item.Name)
	if assert.Len(t, nodes[0].Children, 1) {
		assert.Equal(t, 2, nodes[0].Children[0].Item.ID)
		if assert.Len(t, nodes[0].Children[0].Children, 1) {
			assert.Equal(t, 3, nodes[0].Children[0].Children[0].Item.ID)
		}
	}

	assert.Equal(t, 4, nodes[1].Item.ID)
	assert.Empty(t, nodes[1].Children)
}

func TestFromItemsCallsAccessorsOncePerItem(t *testing.T) {
	items := []testItem{
		{ID: 1, ParentID: 1},
		{ID: 2, ParentID: 1},
	}
	idCalls := 0
	parentIDCalls := 0

	_, err := FromItems(items, func(item testItem) int {
		idCalls++
		return item.ID
	}, func(item testItem) int {
		parentIDCalls++
		return item.ParentID
	})

	assert.NoError(t, err)
	assert.Equal(t, len(items), idCalls)
	assert.Equal(t, len(items), parentIDCalls)
}

func TestFromItemsRejectsDuplicateIDs(t *testing.T) {
	items := []testItem{
		{ID: 1, ParentID: 1},
		{ID: 1, ParentID: 1},
	}

	_, err := FromItems(items, testItemID, testItemParentID)

	assert.EqualError(t, err, "duplicate ID found: 1")
}

func TestFromItemsRejectsMissingParent(t *testing.T) {
	items := []testItem{{ID: 1, ParentID: 9}}

	_, err := FromItems(items, testItemID, testItemParentID)

	assert.EqualError(t, err, "parent node not found for item with id 9")
}

func TestFromItemsRejectsMissingZeroValueParent(t *testing.T) {
	items := []testItem{{ID: 1, ParentID: 0}}

	var nodes []*Node[testItem]
	var err error
	assert.NotPanics(t, func() {
		nodes, err = FromItems(items, testItemID, testItemParentID)
	})

	assert.Nil(t, nodes)
	assert.EqualError(t, err, "parent node not found for item with id 0")
}

func TestFromItemsRejectsCycle(t *testing.T) {
	items := []testItem{
		{ID: 1, ParentID: 2},
		{ID: 2, ParentID: 1},
	}

	_, err := FromItems(items, testItemID, testItemParentID)

	assert.EqualError(t, err, "cycle detected involving item id 1")
}

func TestFromItemsRejectsEmptyItems(t *testing.T) {
	_, err := FromItems([]testItem{}, testItemID, testItemParentID)

	assert.EqualError(t, err, "no root node found (an item where parentId = Id)")
}
