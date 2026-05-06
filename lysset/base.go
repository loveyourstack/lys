package lysset

// Set is a simple set implementation using a map with empty struct values.
type Set[T comparable] map[T]struct{}

// constructors
// -----------------------------------------------------------------------

// New returns a set containing the supplied values.
func New[T comparable](vals ...T) Set[T] {
	s := make(Set[T], len(vals))
	for _, v := range vals {
		s[v] = struct{}{}
	}
	return s
}

// FromSlice returns a set built from vals.
func FromSlice[T comparable](vals []T) Set[T] {
	s := make(Set[T], len(vals))
	for _, v := range vals {
		s[v] = struct{}{}
	}
	return s
}

// methods
// -----------------------------------------------------------------------

// Add inserts val into the set.
func (s *Set[T]) Add(val T) {
	if *s == nil {
		*s = make(Set[T])
	}
	(*s)[val] = struct{}{}
}

// AddAll inserts each supplied value into the set.
func (s *Set[T]) AddAll(vals ...T) {
	if *s == nil {
		*s = make(Set[T], len(vals))
	}
	for _, v := range vals {
		(*s)[v] = struct{}{}
	}
}

// Clear removes all values from the set.
func (s *Set[T]) Clear() {
	if *s == nil {
		return
	}
	for k := range *s {
		delete(*s, k)
	}
}

// Clone returns a copy of the set.
func (s Set[T]) Clone() Set[T] {
	if s == nil {
		return nil
	}
	cpy := make(Set[T], len(s))
	for v := range s {
		cpy[v] = struct{}{}
	}
	return cpy
}

// Contains returns true if val exists in the set.
func (s Set[T]) Contains(val T) bool {
	_, ok := s[val]
	return ok
}

// Equals returns true when both sets have the same values.
func (s Set[T]) Equals(other Set[T]) bool {
	if len(s) != len(other) {
		return false
	}
	for v := range s {
		if !other.Contains(v) {
			return false
		}
	}
	return true
}

// Len returns the number of values in the set.
func (s Set[T]) Len() int {
	return len(s)
}

// Remove deletes val from the set.
func (s *Set[T]) Remove(val T) {
	if *s == nil {
		return
	}
	delete(*s, val)
}

// Values returns the set members in an unspecified order.
func (s Set[T]) Values() []T {
	vals := make([]T, 0, len(s))
	for v := range s {
		vals = append(vals, v)
	}
	return vals
}
