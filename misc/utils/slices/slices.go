package slices

import (
	"fmt"
	"strings"

	"github.com/ryanreadbooks/whimer/misc/generics"
)

// [1,2,3,4,5] -> "1,2,3,4,5"
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

func JoinStrings(strs []string) string {
	return strings.Join(strs, ",")
}

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
