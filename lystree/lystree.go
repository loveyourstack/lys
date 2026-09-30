package lystree

import (
	"fmt"
)

// Node represents a node in a hierarchical tree structure.
// The 'embed' modifier removes the object key when marshaling the Item field to JSON.
type Node[T any] struct {
	Item     T          `json:",embed"`
	Children []*Node[T] `json:"children,omitempty"`
}

// FromItems constructs a hierarchical tree structure from a flat list of items (database structs).
// The idOf function returns the unique identifier of an item.
// The parentIdOf function returns the identifier of the item's parent.
// Root items are identified as those whose parent ID is equal to their own ID.
func FromItems[T any, keyT comparable](
	items []T,
	idOf func(item T) keyT,
	parentIdOf func(item T) keyT,
) (nodes []*Node[T], err error) {

	ids := make([]keyT, len(items))
	parentIDs := make([]keyT, len(items))

	// create map of all items as nodes, keyed by id
	nodeMap := make(map[keyT]*Node[T], len(items))

	// create map of k = id, v = parentId
	parentMap := make(map[keyT]keyT, len(items))

	for idx, item := range items {
		id := idOf(item)
		parentID := parentIdOf(item)

		// check for duplicate IDs
		if _, exists := parentMap[id]; exists {
			return nil, fmt.Errorf("duplicate ID found: %v", id)
		}

		ids[idx] = id
		parentIDs[idx] = parentID

		nodeMap[id] = &Node[T]{
			Item:     item,
			Children: []*Node[T]{},
		}

		parentMap[id] = parentID
	}

	// check for cycles (e.g. 2 -> 3 and 3 -> 2) and missing parent nodes
	for _, start := range ids {
		current := start
		visited := make(map[keyT]struct{})

		for {
			parent, ok := parentMap[current]
			if !ok {
				return nil, fmt.Errorf("parent node not found for item with id %v", current)
			}

			if parent == current {
				break // valid root
			}

			if _, seen := visited[current]; seen {
				return nil, fmt.Errorf("cycle detected involving item id %v", current)
			}

			visited[current] = struct{}{}
			current = parent
		}
	}

	// write node relationships (attach children to their parents)
	for idx, id := range ids {
		parentId := parentIDs[idx]

		// root items have parent_id equal to their own id
		if parentId == id {
			nodes = append(nodes, nodeMap[id])
			continue
		}

		// add item node as a child of parent node
		parentNode := nodeMap[parentId]
		parentNode.Children = append(parentNode.Children, nodeMap[id])
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no root node found (an item where parentId = Id)")
	}

	return nodes, nil
}
