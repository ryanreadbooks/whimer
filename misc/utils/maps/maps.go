package maps

// 获取map中的所有key
// 注意map的遍历是无序的
func Keys[M ~map[K]V, K comparable, V any](m M) []K {
	var keys = make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// 获取map中所有value
// 注意map的遍历是无序的
func Values[M ~map[K]V, K comparable, V any](m M) []V {
	var vals = make([]V, 0, len(m))
	for _, v := range m {
		vals = append(vals, v)
	}
	return vals
}

func All[M ~map[K]V, K comparable, V any](m M) ([]K, []V) {
	l := len(m)
	var keys = make([]K, 0, l)
	var vals = make([]V, 0, l)
	for k, v := range m {
		keys = append(keys, k)
		vals = append(vals, v)
	}

	return keys, vals
}

func Func[M ~map[K]V, K comparable, V any](m M, f func(k K, v V)) {
	for key, value := range m {
		f(key, value)
	}
}
