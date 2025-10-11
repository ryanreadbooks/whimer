package uuid

import (
	"testing"

	"github.com/google/uuid"
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

	if u1.Compare(EmptyUUID()) != 1 {
		t.Errorf("u1.Compare(EmptyUUID()) != 1, want 1")
	}
}

func TestUUid(t *testing.T) {
	id, _ := uuid.NewV7()
	t.Log(id)

	parsed, err := ParseString(id.String())
	t.Log(err)
	t.Log(parsed)

	t.Log(MaxUUID())
}
