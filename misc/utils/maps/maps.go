package maps

// 获取map中的所有key
// 注意map的遍历是无序的
func Keys[T comparable, P any](m map[T]P) []T {
	var keys = make([]T, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// 获取map中所有value
// 注意map的遍历是无序的
func Values[T comparable, P any](m map[T]P) []P {
	var vals = make([]P, 0, len(m))
	for _, v := range m {
		vals = append(vals, v)
	}
	return vals
}

func Func[T comparable, P any](m map[T]P, f func (k T, v P)) {
	for key, value := range m {
		f(key, value)
	}
}
