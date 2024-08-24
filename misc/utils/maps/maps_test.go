package maps

import (
	"testing"
)

func TestMapFunc(t *testing.T) {
	m := map[int]string{1: "abd", 2: "bcd", 3: "qwer"}
	Func(m, func(k int, v string) {
		t.Logf("k = %v, v = %v\n", k, v)
	})
}