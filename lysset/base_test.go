package lysset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSetAndContainsAndLen(t *testing.T) {
	s := New("a", "b", "a")

	assert.Equal(t, 2, s.Len())
	assert.True(t, s.Contains("a"))
	assert.True(t, s.Contains("b"))
	assert.False(t, s.Contains("c"))
}

func TestFromSliceUniqueness(t *testing.T) {
	s := FromSlice([]int{1, 1, 2, 3, 3, 3})

	assert.Equal(t, 3, s.Len())
	assert.True(t, s.Contains(1))
	assert.True(t, s.Contains(2))
	assert.True(t, s.Contains(3))
}

func TestSetAddOnNilSet(t *testing.T) {
	var s Set[int]

	s.Add(10)

	assert.Equal(t, 1, s.Len())
	assert.True(t, s.Contains(10))
}

func TestSetAddAll(t *testing.T) {
	var s Set[int]

	s.AddAll(1, 2, 2, 3)

	assert.Equal(t, 3, s.Len())
	assert.True(t, s.Contains(1))
	assert.True(t, s.Contains(2))
	assert.True(t, s.Contains(3))
}

func TestSetRemove(t *testing.T) {
	s := New(1, 2, 3)

	s.Remove(2)

	assert.Equal(t, 2, s.Len())
	assert.False(t, s.Contains(2))

	// remove missing key should be a no-op
	s.Remove(99)
	assert.Equal(t, 2, s.Len())
}

func TestSetClear(t *testing.T) {
	s := New("x", "y")

	s.Clear()

	assert.Equal(t, 0, s.Len())
	assert.False(t, s.Contains("x"))
	assert.False(t, s.Contains("y"))
}

func TestSetValues(t *testing.T) {
	s := New("one", "two", "three")

	vals := s.Values()

	assert.ElementsMatch(t, []string{"one", "two", "three"}, vals)
}

func TestSetClone(t *testing.T) {
	original := New(1, 2)
	cloned := original.Clone()

	assert.True(t, original.Equals(cloned))

	cloned.Add(3)
	assert.True(t, cloned.Contains(3))
	assert.False(t, original.Contains(3))
}

func TestSetCloneNil(t *testing.T) {
	var s Set[int]

	cloned := s.Clone()

	assert.Nil(t, cloned)
}

func TestSetEquals(t *testing.T) {
	a := New("a", "b", "c")
	b := New("c", "a", "b")
	c := New("a", "b")

	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
