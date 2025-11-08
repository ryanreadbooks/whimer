package xslice

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ryanreadbooks/whimer/misc/generics"
)

// 将整数slice通过','组合在一起返回一个string
//
// 比如 [1,2,3,4,5] -> "1,2,3,4,5"
func JoinInts[T generics.Integer](ints []T) string {
	var ll = len(ints)
	if ll == 0 {
		return ""
	}

	var builder strings.Builder
	builder.Grow(ll*10 + ll - 1)
	builder.WriteString(fmt.Sprintf("%d", ints[0]))
	for _, in := range ints[1:] {
		builder.WriteByte(',')
		builder.WriteString(fmt.Sprintf("%d", in))
	}

	return builder.String()
}

// 将字符串slice通过','组合在一起返回一个string
func JoinStrings(strs []string) string {
	return strings.Join(strs, ",")
}

// slice去重
//
// 比如 [1,2,2,3,4] -> [1,2,3,4]
func Uniq[T comparable](v []T) []T {
	u := make([]T, 0, len(v))
	e := make(map[T]struct{})
	for _, ele := range v {
		if _, ok := e[ele]; !ok {
			u = append(u, ele)
			e[ele] = struct{}{}
		}
	}
	return u
}

func UniqF[T any, C comparable](v []T, f func(v T) C) []T {
	u := make([]T, 0, len(v))
	e := make(map[C]struct{})
	for _, ele := range v {
		if _, ok := e[f(ele)]; !ok {
			u = append(u, ele)
			e[f(ele)] = struct{}{}
		}
	}
	return u
}

func SplitInts[T generics.Integer](s string, sep string) []T {
	res := strings.Split(s, sep)
	// convert string to T
	var ints []T
	if len(res) == 0 {
		return ints
	}

	var zero T
	for _, part := range res {
		switch any(zero).(type) {
		case int, int8, int16, int32, int64:
			tmp, err := strconv.ParseInt(part, 10, 64)
			if err == nil {
				ints = append(ints, T(tmp))
			}
		case uint, uint8, uint16, uint32, uint64:
			tmp, err := strconv.ParseUint(part, 10, 64)
			if err == nil {
				ints = append(ints, T(tmp))
			}
		}
	}
	return ints
}

// 拼接两个slice
//
// 比如 [1,2,3] + [2,3,4] -> [1,2,3,2,3,4]
func Concat[T any](a, b []T) []T {
	u := make([]T, 0, len(a)+len(b))
	u = append(u, a...)
	u = append(u, b...)
	return u
}

// 拼接两个slice后去重
//
// 比如 [1,2,3] + [2,3,4] -> [1,2,3,4]
func ConcatUniq[T comparable](a, b []T) []T {
	return Uniq(Concat(a, b))
}

func AsMap[T comparable](a []T) map[T]struct{} {
	m := make(map[T]struct{}, len(a))
	for idx := range a {
		m[a[idx]] = struct{}{}
	}

	return m
}

func MakeMap[T any, K comparable](t []T, keyFn func(v T) K) map[K]T {
	var r = make(map[K]T, len(t))
	for _, e := range t {
		r[keyFn(e)] = e
	}

	return r
}

func Repeat[T any](v T, n int) []T {
	r := make([]T, 0, n)
	for range n {
		r = append(r, v)
	}

	return r
}

// filter func returns true means discarding current value
func Filter[V any, T ~[]V](t T, filter func(idx int, v V) bool) T {
	dest := make(T, 0, len(t))
	for idx, v := range t {
		if !filter(idx, v) { // got included
			dest = append(dest, v)
		}
	}
	return dest
}

// filter zero values in slice
func FilterZero[V comparable, T ~[]V](t T) T {
	var z V
	return Filter(t, func(_ int, v V) bool { return v == z })
}

func Any[T any](t []T) []any {
	a := make([]any, 0, len(t))
	for _, item := range t {
		a = append(a, item)
	}

	return a
}
