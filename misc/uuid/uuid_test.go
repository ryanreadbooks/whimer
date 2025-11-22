package uuid

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/ryanreadbooks/whimer/misc/xlog"
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

	t.Log(EmptyUUID().String())
}

func TestMonotonic(t *testing.T) {
	u1 := NewUUID() // smaller
	u2 := NewUUID() // greater
	t.Log(u1)
	t.Log(u2)
	t.Log(u2.Compare(u1))

	uuids := make([]UUID, 0, 1000) // asc order
	for range 1000 {
		uuids = append(uuids, NewUUID())
	}

	// check uuids is asc order
	for i := 0; i < 999; i++ {
		if i%50 == 0 {
			// print timestamp
			t.Logf("%d: %d", i, uuids[i].Time().UnixMilli())
		}
		if uuids[i].Compare(uuids[i+1]) >= 0 {
			t.Errorf("uuid[%d] >= uuid[%d], want less than", i, i+1)
		}
	}
}

func TestUUIDString(t *testing.T) {
	id := NewUUID()
	parsed, err := ParseString(id.String())
	if err != nil {
		t.Errorf("ParseString(%s) failed, err: %v", id.String(), err)
	}
	// id should equal to parsed
	if id.Compare(parsed) != 0 {
		t.Errorf("id != parsed, want true")
	}
	t.Log(id.String())
	t.Log(parsed.String())
	t.Log(parsed.UUID.String())
}

func TestUUIDLog(t *testing.T) {
	uu := NewUUID()
	xlog.Msg("check").Extras("uuid", uu).Info()
	t.Logf("%s\n", uu)
	t.Logf("%v\n", uu)
	t.Logf("%+v\n", uu)
	c, _ := json.Marshal(uu)
	t.Logf("%s\n", c)

	msg := struct {
		MsgId UUID `json:"msg_id"`
	}{}

	texts := []string{
		`{"msg_id": "019aa9a21afc70c6ac63385b68a11af3"}`,
		`{"msg_id": "019a68a5-19fc-7b18-b3b8-3fc08a2d0779"}`,
	}
	for _, text := range texts {
		err := json.Unmarshal([]byte(text), &msg)
		if err != nil {
			t.Fatal(err)
		}

		t.Log(msg.MsgId)
	}
}
