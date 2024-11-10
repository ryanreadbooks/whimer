package csrf

import (
	"testing"
)

func TestGetToken(t *testing.T) {
	n := 5
	for i := 0; i < n; i++ {
		tk := GetToken()
		t.Log(tk, len(tk))
	}
}
