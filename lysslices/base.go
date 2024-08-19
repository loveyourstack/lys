package lysslices

import (
	"cmp"
	"slices"
)

// EqualUnordered true if s1 and s2 are equal regardless of sorting
func EqualUnordered[S ~[]E, E cmp.Ordered](s1, s2 S) bool {

	if len(s1) != len(s2) {
		return false
	}

	s1s := make([]E, len(s1))
	for _, v := range s1 {
		s1s = append(s1s, v)
	}

	s2s := make([]E, len(s2))
	for _, v := range s2 {
		s2s = append(s2s, v)
	}

	slices.Sort(s1s)
	slices.Sort(s2s)
	return slices.Equal(s1s, s2s)
}
