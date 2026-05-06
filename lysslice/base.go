package lysslice

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/loveyourstack/lys/lysset"
)

// ContainsAll returns true if all elements are found in slice. Elements may contain duplicates.
func ContainsAll[T comparable](slice []T, elements []T) bool {

	if len(slice) == 0 || len(elements) == 0 {
		return false
	}

	sliceSet := lysset.FromSlice(slice)

	for _, ele := range elements {
		if !sliceSet.Contains(ele) {
			return false
		}
	}

	return true
}

// ContainsAny returns true if any element is found in slice.
func ContainsAny[T comparable](slice []T, elements []T) bool {

	if len(slice) == 0 || len(elements) == 0 {
		return false
	}

	sliceSet := lysset.FromSlice(slice)

	return slices.ContainsFunc(elements, sliceSet.Contains)
}

// DeDuplicate returns a copy of the slice with duplicates removed and without affecting value order.
func DeDuplicate[S ~[]E, E comparable](s S) S {

	if s == nil {
		return nil
	}
	if len(s) == 0 {
		return s
	}

	seen := lysset.New[E]()
	s1 := make(S, 0, len(s))

	for _, v := range s {
		if !seen.Contains(v) {
			seen.Add(v)
			s1 = append(s1, v)
		}
	}

	return s1
}

// EqualUnordered returns true if s1 and s2 are equal regardless of sorting.
func EqualUnordered[S ~[]E, E comparable](s1, s2 S) bool {

	if len(s1) != len(s2) {
		return false
	}

	freq := make(map[E]int, len(s1))
	for _, v := range s1 {
		freq[v]++
	}

	for _, v := range s2 {
		count, ok := freq[v]
		if !ok {
			return false
		}
		if count == 1 {
			delete(freq, v)
		} else {
			freq[v] = count - 1
		}
	}

	return len(freq) == 0
}

func ReportDuplicates[S ~[]E, E comparable](s S) (dups []string) {

	if len(s) == 0 {
		return nil
	}

	dups = make([]string, 0, len(s))

	seen := make(map[E]int, len(s))
	for _, v := range s {
		seen[v]++
	}

	for v, count := range seen {
		if count > 1 {
			dups = append(dups, fmt.Sprintf("%v", v))
		}
	}

	if len(dups) == 0 {
		return nil
	}

	slices.Sort(dups)
	return dups
}

// SortAndDeDuplicate returns a sorted and de-duplicated copy of the slice.
func SortAndDeDuplicate[S ~[]E, E cmp.Ordered](s S) S {

	if s == nil {
		return nil
	}
	if len(s) == 0 {
		return s
	}

	s1 := slices.Clone(s)
	slices.Sort(s1)
	return slices.Compact(s1)
}
