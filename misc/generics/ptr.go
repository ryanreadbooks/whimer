package generics

func Ptr[T any](a T) *T {
	return &a
}

func Bool[T Integer](t T) bool {
	return t == 1
}

func FromBool[T Integer](b bool) T {
	if b {
		return 1
	}
	return 0
}
