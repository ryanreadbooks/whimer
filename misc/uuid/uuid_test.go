package uuid

import (
	"testing"
)

func TestUUIDCompare(t *testing.T) {
	u1 := NewUUID()
	u2 := NewUUID()
	if u1 == u2 {	
		t.Errorf("u1 == u2, want false")
	}

	if !u1.LessThan(u2) {
		t.Errorf("u1.LessThan(u2) == false, want true")
	}

	u3 := u1.Duplicate()
	if u3 != u1 {
		t.Errorf("u3 != u1, want true")
	}

	if u1.Compare(u2) == 0 {
		t.Errorf("u1.Compare(u2) == 0, want not 0")
	}
}