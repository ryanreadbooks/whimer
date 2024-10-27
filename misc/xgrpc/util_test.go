package xgrpc

import (
	"strings"
	"testing"
)

func TestSplitN(t *testing.T) {
	seps := strings.SplitN("/note.sdk.v1.NoteService/IsUserOwnNote", "/", 3)
	for i, s := range seps {
		t.Log(i, s)
	}
	seps = strings.SplitN("helloworld", "/", 3)
	for i, s := range seps {
		t.Log(i, s)
	}
}
