package svc

import (
	"fmt"
	"math/bits"
	"strconv"
	"testing"
)

func fk1(oid uint64, biz int) string {
	return fmt.Sprintf("summary:%d:%d", biz, oid)
}

func fk2(oid uint64, biz int) string {
	return "summary:" + strconv.Itoa(biz) + ":" + strconv.FormatUint(oid, 10)
}

func BenchmarkFormKeyFk1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fk1(2392923, 10002)
	}
}

func BenchmarkFormKeyFk2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fk2(2392923, 10002)
	}
}

func TestAdd(t *testing.T) {
	num, underflow := bits.Sub64(1, 2, 0)
	t.Log(num, underflow)
	num, underflow = bits.Sub64(0, 2, 0)
	t.Log(num, underflow)
}
