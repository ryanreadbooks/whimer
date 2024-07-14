package safety

import (
	"testing"
)

func TestConfuseUnsigned(t *testing.T) {
	s := Confuse(18)
	t.Log(s)

	n := DeConfuse(s)
	t.Log(n)
}

func TestConfuseUint64(t *testing.T) {
	confuser := NewConfuser("0x7c00:noteIdConfuser:.$35%io", 24)
	t.Log(confuser.ConfuseU(8))
	t.Log(confuser.DeConfuseU("g0qyNoE5Y9VZRValbeLdmkjr"))
}