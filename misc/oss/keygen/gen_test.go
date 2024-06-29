package keygen

import (
	"testing"
)

func TestKeyGen(t *testing.T) {
	gen := NewGenerator(WithBucket("test-bucket"))
	t.Log(gen.Gen())
	gen2 := NewGenerator(WithBucket("test-bucket"), WithPrefix("misc"))
	t.Log(gen2.Gen())
}
