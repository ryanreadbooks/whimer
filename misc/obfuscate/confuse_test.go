package obfuscate

import (
	"testing"
)

func TestConfuseUnsigned(t *testing.T) {
	s, _ := Mix(18)
	t.Log(s)

	n, _ := DeMix(s)
	t.Log(n)
}

func TestConfuseUint64(t *testing.T) {
	confuser, _ := NewConfuser(WithSalt("0x7c00:noteIdConfuser:.$35%io"),
		WithMinLen(10))
	t.Log(confuser.MixU(8))
	t.Log(confuser.DeMixU("5Y9VZRValb"))
}
