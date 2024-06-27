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
