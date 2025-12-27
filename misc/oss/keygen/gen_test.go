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

func TestKeyGenUnwrap(t *testing.T) {
	gen := NewGenerator(WithBucket("test-bucket"), WithPrefix("misc"))
	genId := gen.Gen()
	t.Log(genId)
	t.Log(gen.Check("test-bucket/misc/1234567890"))
	t.Log(gen.Check("1234567890"))
	t.Log(gen.Check("nota/1234567890"))
	t.Log(gen.Check("test-bucket/prefix/1234567890"))
	t.Log(gen.Check("test-bucket/1234567890"))
	t.Log(gen.Check("test-bucket/misc/1234567890"))
	t.Log(gen.Check("test-bucket/misc/1234567890"))

	gen2 := NewGenerator(WithBucket("test-bucket"), WithPrefix("misc/1"), WithPrependPrefix(true))
	gen2Id := gen2.Gen()
	t.Log(gen2Id)
	t.Log(gen2.Check("test-bucket/misc/1234567890"))   //f
	t.Log(gen2.Check("1234567890"))                    //f
	t.Log(gen2.Check("nota/1234567890"))               //f
	t.Log(gen2.Check("test-bucket/prefix/1234567890")) //f
	t.Log(gen2.Check("test-bucket/1234567890"))        //f
	t.Log(gen2.Check("test-bucket/misc/1/1234567890")) //t
}
