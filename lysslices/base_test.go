package lysslices

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortEqualSuccess(t *testing.T) {
	s1 := []int{1, 2, 3}
	s2 := []int{3, 1, 2}
	eq := SortEqual(s1, s2)
	assert.Equal(t, true, eq)

	sa := []string{"a", "b", "c"}
	sb := []string{"c", "a", "b"}
	eq = SortEqual(sa, sb)
	assert.Equal(t, true, eq)
}

func TestSortEqualFailure(t *testing.T) {
	s1 := []int{1, 2, 3}
	s2 := []int{1, 2}
	eq := SortEqual(s1, s2)
	assert.Equal(t, false, eq)

	sa := []string{"a", "b", "c"}
	sb := []string{"a", "b"}
	eq = SortEqual(sa, sb)
	assert.Equal(t, false, eq)
}
