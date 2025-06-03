package xtime

import (
	"encoding/json"
	"testing"
)

func TestXDuration(t *testing.T) {
	b := `{"timeout": "1m"}`
	s := struct {
		Timeout Duration `json:"timeout"`
	}{}
	err := json.Unmarshal([]byte(b), &s)
	t.Log(err)
	t.Log(s)
}
