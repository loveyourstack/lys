package lysslices

import (
	"cmp"
	"slices"
)

// SortEqual sorts s1 and s2 then returns true if they are equal
func SortEqual[S ~[]E, E cmp.Ordered](s1, s2 S) bool {
	slices.Sort(s1)
	slices.Sort(s2)
	return slices.Equal(s1, s2)
}
