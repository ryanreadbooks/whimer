package xslice

import (
	"fmt"
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

func Repeat[T any](v T, n int) []T {
	r := make([]T, 0, n)
	for range n {
		r = append(r, v)
	}

	return r
}
