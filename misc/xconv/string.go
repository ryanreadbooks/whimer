package xconv

import (
	"strconv"

	"github.com/ryanreadbooks/whimer/misc/generics"
)

func FormatUint[T generics.UnSignedInteger](u T) string {
	return strconv.FormatUint(uint64(u), 10)
}

func FormatInt[T generics.SignedInteger](s T) string {
	return strconv.FormatInt(int64(s), 10)
}
