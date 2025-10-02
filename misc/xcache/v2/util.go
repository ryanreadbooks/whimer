package v2

func Keys[T any](t []T, keyGen func(T) string) []string {
	l := len(t)
	keys := make([]string, 0, l)
	for _, item := range t {
		key := keyGen(item)
		keys = append(keys, key)
	}

	return keys
}

func KeysAndMap[T any](t []T, keyGen func(T) string) ([]string, map[string]T) {
	l := len(t)
	keys := make([]string, 0, l)
	maps := make(map[string]T, l)
	for _, item := range t {
		key := keyGen(item)
		keys = append(keys, key)
		maps[key] = item
	}

	return keys, maps
}

// 遍历keys
// 从m中找到遍历的key 并回调revert中
func RangeRevertKeys[T any, R any](keys []string, m map[string]T, revert func(T) R) []R {
	rs := make([]R, 0, len(keys))
	for _, k := range keys {
		rs = append(rs, revert(m[k]))
	}

	return rs
}
