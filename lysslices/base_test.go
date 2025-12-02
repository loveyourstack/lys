package lysslices

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsAllSuccess(t *testing.T) {
	s := []int{1, 2}

	e1 := []int{1}
	assert.Equal(t, true, ContainsAll(s, e1), "e1")

	e2 := []int{1, 2}
	assert.Equal(t, true, ContainsAll(s, e2), "e2")
}

func TestContainsAllFailure(t *testing.T) {
	s := []int{1, 2}

	e1 := []int{}
	assert.Equal(t, false, ContainsAll(s, e1), "e1")

	e2 := []int{1, 3}
	assert.Equal(t, false, ContainsAll(s, e2), "e2")

	e3 := []int{1, 2, 3}
	assert.Equal(t, false, ContainsAll(s, e3), "e3")

	s1 := []int{}
	assert.Equal(t, false, ContainsAll(s1, e2), "s1")
}

func TestDeDuplicateSuccess(t *testing.T) {
	s1 := []int{2, 3, 1, 3, 2}
	s1a := DeDuplicate(s1)
	assert.Equal(t, []int{2, 3, 1, 3, 2}, s1)
	assert.Equal(t, []int{2, 3, 1}, s1a)

	s2 := []string{"b", "c", "a", "b", "a"}
	s2a := DeDuplicate(s2)
	assert.Equal(t, []string{"b", "c", "a", "b", "a"}, s2)
	assert.Equal(t, []string{"b", "c", "a"}, s2a)
}

func TestEqualUnorderedSuccess(t *testing.T) {
	s1 := []int{1, 2, 3}
	s2 := []int{3, 1, 2}
	eq := EqualUnordered(s1, s2)
	assert.Equal(t, true, eq)

	sa := []string{"a", "b", "c"}
	sb := []string{"c", "a", "b"}
	eq = EqualUnordered(sa, sb)
	assert.Equal(t, true, eq)
}

func TestEqualUnorderedFailure(t *testing.T) {
	s1 := []int{1, 2, 3}
	s2 := []int{1, 2}
	eq := EqualUnordered(s1, s2)
	assert.Equal(t, false, eq)

	s3 := []int{1, 2, 3}
	s4 := []int{1, 2, 4}
	eq = EqualUnordered(s3, s4)
	assert.Equal(t, false, eq)

	sa := []string{"a", "b", "c"}
	sb := []string{"a", "b"}
	eq = EqualUnordered(sa, sb)
	assert.Equal(t, false, eq)

	sc := []string{"a", "b", "c"}
	sd := []string{"a", "b", "d"}
	eq = EqualUnordered(sc, sd)
	assert.Equal(t, false, eq)
}

func TestSortAndDeDuplicateSuccess(t *testing.T) {
	s1 := []int{2, 3, 1, 3, 2}
	s1a := SortAndDeDuplicate(s1)
	assert.Equal(t, []int{2, 3, 1, 3, 2}, s1)
	assert.Equal(t, []int{1, 2, 3}, s1a)

	s2 := []string{"b", "c", "a", "b", "a"}
	s2a := SortAndDeDuplicate(s2)
	assert.Equal(t, []string{"b", "c", "a", "b", "a"}, s2)
	assert.Equal(t, []string{"a", "b", "c"}, s2a)
}
