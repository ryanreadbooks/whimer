package signinup

import (
	"testing"
)

func TestMakeSalt(t *testing.T) {
	salt, err := makeSalt()
	t.Log(err)
	t.Log(salt)
}

func TestMakeInitPass(t *testing.T) {
	pass, salt, err := makeInitPass()
	t.Log(pass)
	t.Log(salt)
	t.Log(err)
}
