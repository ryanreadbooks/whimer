package generics

func Ptr[T any](a T) *T {
	return &a
}
