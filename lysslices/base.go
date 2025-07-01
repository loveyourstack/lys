package lysslices

import (
	"cmp"
	"slices"
)

// ContainsAny returns true if any string in elements is found in slice
func ContainsAny[T comparable](slice []T, elements []T) bool {
	for _, v := range elements {
		if slices.Contains(slice, v) {
			return true
		}
	}
	return false
}

// DeDuplicate returns a copy of the slice with duplicates removed and without affecting value order
// from https://stackoverflow.com/questions/66643946/how-to-remove-duplicates-strings-or-int-from-slice-in-go
func DeDuplicate[S ~[]E, E comparable](s S) S {
	keys := make(map[E]bool)
	s1 := []E{}
	for _, v := range s {
		if _, ok := keys[v]; !ok {
			keys[v] = true
			s1 = append(s1, v)
		}
	}
	return s1
}

// EqualUnordered true if s1 and s2 are equal regardless of sorting
func EqualUnordered[S ~[]E, E cmp.Ordered](s1, s2 S) bool {

	if len(s1) != len(s2) {
		return false
	}

	s1s := make([]E, len(s1))
	for i := range s1 {
		s1s[i] = s1[i]
	}

	s2s := make([]E, len(s2))
	for i := range s2 {
		s2s[i] = s2[i]
	}

	slices.Sort(s1s)
	slices.Sort(s2s)
	return slices.Equal(s1s, s2s)
}

// SortAndDeDuplicate returns a sorted and de-duplicated copy of the slice
func SortAndDeDuplicate[S ~[]E, E cmp.Ordered](s S) S {

	s1 := make([]E, len(s))
	for i := range s {
		s1[i] = s[i]
	}

	slices.Sort(s1)
	return slices.Compact(s1)
}
