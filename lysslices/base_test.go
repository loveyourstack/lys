package lysslices

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
