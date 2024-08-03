package maps

// 获取map中的所有key
func Keys[T comparable, P any](m map[T]P) []T {
	var keys = make([]T, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
