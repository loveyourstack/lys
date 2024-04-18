package lystype

func ToPtr[T any](a T) *T {
	return &a
}
