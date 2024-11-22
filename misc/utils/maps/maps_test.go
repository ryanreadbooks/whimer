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

func TestMapBatchExec(t *testing.T) {
	m := map[int]string{1: "abd", 2: "bcd", 3: "qwer", 10: "12123", 123: "cnasdlkf"}
	BatchExec(m, 0, func(target map[int]string) error {
		t.Log(target)
		return nil
	})
}
